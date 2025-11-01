// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"

	"github.com/lvcasx1/ithil/pkg/types"
)

// Config holds the Telegram client configuration.
type Config struct {
	APIID            string
	APIHash          string
	SessionFile      string
	DatabaseDirectory string
}

// Client wraps the gotd Telegram client.
type Client struct {
	config  *Config
	logger  *slog.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	updates chan *types.Update
	authStateChanges chan types.AuthState

	// API credentials
	apiID   int
	apiHash string

	// gotd client and API
	client    *telegram.Client
	api       *tg.Client
	sessionStorage *SessionStorage

	// Authentication state management
	authMu        sync.RWMutex
	authState     types.AuthState
	isAuthenticated bool
	currentUser   *tg.User

	// Update handler control
	updateHandler *UpdateHandler
	gaps          *updates.Manager
	updatesStarted bool
	wg           sync.WaitGroup
}

// New creates a new Telegram client.
func New(config *Config, logger *slog.Logger, ctx context.Context) (*Client, error) {
	// Validate config
	if config.APIID == "" || config.APIHash == "" {
		return nil, fmt.Errorf("API ID and API Hash must be provided")
	}

	// Parse API ID to integer
	apiID, err := strconv.Atoi(config.APIID)
	if err != nil {
		return nil, fmt.Errorf("invalid API ID: %w", err)
	}

	// Create context with cancellation
	clientCtx, cancel := context.WithCancel(ctx)

	// Create session storage
	sessionStorage := NewSessionStorage(config.SessionFile)

	client := &Client{
		config:           config,
		logger:           logger,
		ctx:              clientCtx,
		cancel:           cancel,
		updates:          make(chan *types.Update, 100),
		authStateChanges: make(chan types.AuthState, 10),
		apiID:            apiID,
		apiHash:          config.APIHash,
		sessionStorage:   sessionStorage,
		authState:        types.AuthStateWaitPhoneNumber,
		isAuthenticated:  false,
	}

	// Create update handler (cache will be set later via SetCache)
	client.updateHandler = NewUpdateHandler(client, nil, logger)

	// Create updates manager with our handler
	// This manages the gap recovery and update state persistence
	client.gaps = updates.New(updates.Config{
		Handler: client.updateHandler,
	})

	logger.Info("Created gaps update manager")

	// Create gotd client options with required device/app information
	opts := telegram.Options{
		SessionStorage: sessionStorage,
		UpdateHandler:  client.gaps, // CRITICAL: Pass gaps as the update handler
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(client.gaps.Handle), // CRITICAL: Add middleware to route updates to gaps
		},
		Device: telegram.DeviceConfig{
			DeviceModel:    "Desktop",
			SystemVersion:  "Linux",
			AppVersion:     "0.1.0",
			SystemLangCode: "en",
			LangPack:       "tdesktop",
			LangCode:       "en",
		},
	}

	// Create the telegram client
	client.client = telegram.NewClient(apiID, config.APIHash, opts)
	client.api = client.client.API()

	return client, nil
}

// Start starts the Telegram client and connects to Telegram.
func (c *Client) Start() error {
	c.logger.Info("Starting Telegram client...")

	// Run the client in a goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		if err := c.client.Run(c.ctx, func(ctx context.Context) error {
			// Try to check authentication status
			status, err := c.client.Auth().Status(ctx)

			// Handle auth status errors gracefully
			if err != nil {
				c.logger.Warn("Could not get auth status, assuming not authenticated", "error", err)

				// Check if this is an invalid session error - clear it
				errStr := err.Error()
				if strings.Contains(errStr, "AUTH_KEY_UNREGISTERED") ||
				   strings.Contains(errStr, "SESSION_REVOKED") ||
				   strings.Contains(errStr, "AUTH_KEY_DUPLICATED") {
					c.logger.Warn("Invalid session detected, clearing session...")
					c.sessionStorage.ClearSession()
					c.sessionStorage.ClearAuthData()
				}

				// Set to unauthenticated state and keep running
				c.authMu.Lock()
				c.isAuthenticated = false
				c.authState = types.AuthStateWaitPhoneNumber
				c.authMu.Unlock()

				// Notify UI of auth state
				select {
				case c.authStateChanges <- types.AuthStateWaitPhoneNumber:
				default:
				}

				c.logger.Info("Telegram client connected", "authenticated", false)

				// Wait for context cancellation (allow auth to happen)
				<-ctx.Done()
				return nil
			}

			// Auth status retrieved successfully
			c.authMu.Lock()
			c.isAuthenticated = status.Authorized
			var authState types.AuthState
			if status.Authorized {
				c.authState = types.AuthStateReady
				authState = types.AuthStateReady
				// Get current user info
				user, err := c.api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
				if err == nil {
					if u, ok := user.Users[0].(*tg.User); ok {
						c.currentUser = u
					}
				}
			} else {
				c.authState = types.AuthStateWaitPhoneNumber
				authState = types.AuthStateWaitPhoneNumber
			}
			c.authMu.Unlock()

			// Notify UI of auth state
			select {
			case c.authStateChanges <- authState:
			default:
			}

			c.logger.Info("Telegram client connected", "authenticated", status.Authorized)

			// If authenticated, start gaps.Run() to begin receiving updates
			if status.Authorized {
				c.authMu.Lock()
				if !c.updatesStarted {
					c.updatesStarted = true
					// Get user ID for updates
					userID := int64(0)
					if c.currentUser != nil {
						userID = c.currentUser.ID
					}
					c.authMu.Unlock()

					c.logger.Info("=== STARTING GAPS.RUN() ===", "userID", userID)

					// CRITICAL: Call gaps.Run() to start the update loop
					// This MUST be called within the client.Run() context
					// gaps.Run() will block until context is done, receiving and dispatching updates
					if err := c.gaps.Run(ctx, c.api, userID, updates.AuthOptions{
						OnStart: func(ctx context.Context) {
							c.logger.Info("!!! GAPS STARTED - UPDATES SHOULD NOW FLOW !!!")
						},
					}); err != nil {
						c.logger.Error("Update handler error", "error", err)
						return err
					}

					c.logger.Info("=== GAPS.RUN() STOPPED ===")
					return nil
				} else {
					c.authMu.Unlock()
					c.logger.Info("Updates already started, skipping gaps.Run()")
				}
			}

			// Wait for context cancellation
			<-ctx.Done()
			return nil
		}); err != nil {
			c.logger.Error("Telegram client error", "error", err)
		}
	}()

	return nil
}

// Stop stops the Telegram client.
func (c *Client) Stop() error {
	c.logger.Info("Stopping Telegram client...")

	// Cancel the context
	c.cancel()

	// Wait for goroutines to finish
	c.wg.Wait()

	// Close channels
	close(c.updates)
	close(c.authStateChanges)

	c.logger.Info("Telegram client stopped")
	return nil
}

// Updates returns the channel for receiving updates.
func (c *Client) Updates() <-chan *types.Update {
	return c.updates
}

// AuthStateChanges returns the channel for receiving auth state changes.
func (c *Client) AuthStateChanges() <-chan types.AuthState {
	return c.authStateChanges
}

// IsAuthenticated checks if the client is authenticated.
func (c *Client) IsAuthenticated() bool {
	c.authMu.RLock()
	defer c.authMu.RUnlock()
	return c.isAuthenticated
}

// GetAuthState returns the current authentication state.
func (c *Client) GetAuthState() types.AuthState {
	c.authMu.RLock()
	defer c.authMu.RUnlock()
	return c.authState
}

// GetCurrentUserID returns the current user's ID.
func (c *Client) GetCurrentUserID() int64 {
	c.authMu.RLock()
	defer c.authMu.RUnlock()
	if c.currentUser != nil {
		return c.currentUser.ID
	}
	return 0
}

// ClearSession clears the session and auth data (for invalid/corrupted sessions).
func (c *Client) ClearSession() error {
	c.logger.Info("Clearing session and auth data...")
	if err := c.sessionStorage.ClearSession(); err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}
	if err := c.sessionStorage.ClearAuthData(); err != nil {
		return fmt.Errorf("failed to clear auth data: %w", err)
	}
	c.authMu.Lock()
	c.isAuthenticated = false
	c.authState = types.AuthStateWaitPhoneNumber
	c.authMu.Unlock()
	c.logger.Info("Session cleared successfully")
	return nil
}

// setAuthState updates the authentication state.
func (c *Client) setAuthState(state types.AuthState) {
	c.authMu.Lock()
	c.authState = state
	wasAuthenticated := c.isAuthenticated
	c.isAuthenticated = (state == types.AuthStateReady)
	c.authMu.Unlock()

	// If we just became authenticated and updates aren't started, start them
	if !wasAuthenticated && c.isAuthenticated && !c.updatesStarted {
		c.startUpdateHandler()
	}

	// Notify listeners of auth state change (non-blocking)
	select {
	case c.authStateChanges <- state:
	default:
		c.logger.Warn("Auth state change channel full, skipping notification")
	}
}

// startUpdateHandler starts the update handler for real-time updates.
// NOTE: This is called from the auth flow when authentication completes.
// Since gaps.Run() must be called from within client.Run(), this method
// is actually NOT used anymore - updates are started automatically by
// the middleware and UpdateHandler integration in client options.
func (c *Client) startUpdateHandler() {
	c.authMu.Lock()
	if c.updatesStarted {
		c.authMu.Unlock()
		c.logger.Info("Updates already started, skipping")
		return
	}
	c.updatesStarted = true
	userID := int64(0)
	if c.currentUser != nil {
		userID = c.currentUser.ID
	}
	c.authMu.Unlock()

	c.logger.Info("Update handler setup triggered from auth flow", "userID", userID)
	c.logger.Info("Note: gaps.Run() will be called automatically within client.Run() context")
}

// SetCache sets the cache for the update handler.
// This should be called after creating the cache in the main model.
func (c *Client) SetCache(cache interface {
	SetUser(user *types.User)
	AddMessage(chatID int64, message *types.Message)
}) {
	if c.updateHandler != nil {
		c.updateHandler.SetCache(cache)
		c.logger.Info("Cache set on update handler")
	}
}
