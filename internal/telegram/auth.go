// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"

	"github.com/lvcasx1/ithil/pkg/types"
)

// authFlow implements the auth.FlowClient interface for interactive authentication.
type authFlow struct {
	client        *Client
	phoneNumber   string
	phoneCodeHash string
	firstName     string
	lastName      string
}

// SendPhoneNumber sends the phone number for authentication.
func (c *Client) SendPhoneNumber(phoneNumber string) error {
	c.logger.Info("Sending phone number for authentication", "phone", phoneNumber)

	// Check if already authenticated - prevent AUTH_RESTART error
	if c.IsAuthenticated() {
		c.logger.Warn("Already authenticated, ignoring phone number request")
		return errors.New("already authenticated - session is active")
	}

	// Clean up phone number (remove spaces, dashes, etc.)
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")

	// Ensure phone number starts with +
	if !strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = "+" + phoneNumber
	}

	// Send the code request to Telegram
	sentCode, err := c.api.AuthSendCode(c.ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phoneNumber,
		APIID:       c.apiID,
		APIHash:     c.apiHash,
		Settings:    tg.CodeSettings{},
	})

	if err != nil {
		// Check if this is an AUTH_RESTART error (invalid/corrupted session)
		if strings.Contains(err.Error(), "AUTH_RESTART") {
			c.logger.Warn("Invalid session detected, clearing session and retrying...")
			// Clear the corrupted session
			if clearErr := c.sessionStorage.ClearSession(); clearErr != nil {
				c.logger.Error("Failed to clear session", "error", clearErr)
			}
			if clearErr := c.sessionStorage.ClearAuthData(); clearErr != nil {
				c.logger.Error("Failed to clear auth data", "error", clearErr)
			}
			return errors.New("invalid session detected - please restart the application to authenticate again")
		}

		c.logger.Error("Failed to send code", "error", err)
		return fmt.Errorf("failed to send code: %w", err)
	}

	// Extract phone code hash
	var phoneCodeHash string
	switch code := sentCode.(type) {
	case *tg.AuthSentCode:
		phoneCodeHash = code.PhoneCodeHash
	case *tg.AuthSentCodeSuccess:
		// Already authorized somehow
		c.setAuthState(types.AuthStateReady)
		return nil
	default:
		return fmt.Errorf("unexpected auth response type: %T", sentCode)
	}

	// Save auth data for later use
	authData := &AuthData{
		PhoneNumber:   phoneNumber,
		PhoneCodeHash: phoneCodeHash,
		IsRegistered:  true, // We'll find out if false later
	}

	if err := c.sessionStorage.SaveAuthData(authData); err != nil {
		return fmt.Errorf("failed to save auth data: %w", err)
	}

	// Update state to wait for code
	c.setAuthState(types.AuthStateWaitCode)

	c.logger.Info("Code sent successfully")
	return nil
}

// SendAuthCode sends the authentication code.
func (c *Client) SendAuthCode(code string) error {
	c.logger.Info("Sending authentication code")

	// Clean up the code (remove spaces, dashes, etc.)
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")

	if code == "" {
		return errors.New("code cannot be empty")
	}

	c.logger.Info("Cleaned code", "length", len(code))

	// Load auth data to get phone code hash
	authData, err := c.sessionStorage.LoadAuthData()
	if err != nil {
		return fmt.Errorf("failed to load auth data: %w", err)
	}
	if authData == nil || authData.PhoneCodeHash == "" {
		return errors.New("no pending authentication request")
	}

	c.logger.Info("Using auth data", "phone", authData.PhoneNumber, "has_hash", authData.PhoneCodeHash != "")

	// Sign in with the code
	_, err = c.api.AuthSignIn(c.ctx, &tg.AuthSignInRequest{
		PhoneNumber:   authData.PhoneNumber,
		PhoneCodeHash: authData.PhoneCodeHash,
		PhoneCode:     code,
	})

	if err != nil {
		// Check if this is a 2FA error
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
			c.setAuthState(types.AuthStateWaitPassword)
			return nil
		}

		// Check if user needs to register
		if strings.Contains(err.Error(), "PHONE_NUMBER_UNOCCUPIED") {
			authData.IsRegistered = false
			c.sessionStorage.SaveAuthData(authData)
			c.setAuthState(types.AuthStateWaitRegistration)
			return nil
		}

		// Provide better error messages
		if strings.Contains(err.Error(), "PHONE_CODE_INVALID") {
			c.logger.Error("Invalid code", "error", err)
			return errors.New("invalid verification code - please check the code and try again")
		}

		if strings.Contains(err.Error(), "PHONE_CODE_EXPIRED") {
			c.logger.Error("Code expired", "error", err)
			return errors.New("verification code expired - please request a new code")
		}

		c.logger.Error("Failed to sign in with code", "error", err)
		return fmt.Errorf("failed to verify code: %w", err)
	}

	// Authentication successful
	c.setAuthState(types.AuthStateReady)

	// Get current user info
	user, err := c.api.UsersGetFullUser(c.ctx, &tg.InputUserSelf{})
	if err == nil {
		if u, ok := user.Users[0].(*tg.User); ok {
			c.authMu.Lock()
			c.currentUser = u
			c.authMu.Unlock()
		}
	}

	// Clear temporary auth data
	c.sessionStorage.ClearAuthData()

	c.logger.Info("Authentication successful - real-time updates enabled")
	return nil
}

// SendPassword sends the 2FA password.
func (c *Client) SendPassword(password string) error {
	c.logger.Info("Sending 2FA password")

	// Use the auth client to verify password
	_, err := c.client.Auth().Password(c.ctx, password)
	if err != nil {
		c.logger.Error("Failed to verify password", "error", err)
		return fmt.Errorf("incorrect password: %w", err)
	}

	// Authentication successful
	c.setAuthState(types.AuthStateReady)

	// Get current user info
	user, err := c.api.UsersGetFullUser(c.ctx, &tg.InputUserSelf{})
	if err == nil {
		if u, ok := user.Users[0].(*tg.User); ok {
			c.authMu.Lock()
			c.currentUser = u
			c.authMu.Unlock()
		}
	}

	// Clear temporary auth data
	c.sessionStorage.ClearAuthData()

	c.logger.Info("2FA authentication successful")
	return nil
}

// SendRegistration sends registration information.
func (c *Client) SendRegistration(firstName, lastName string) error {
	c.logger.Info("Sending registration information", "firstName", firstName, "lastName", lastName)

	// Load auth data
	authData, err := c.sessionStorage.LoadAuthData()
	if err != nil {
		return fmt.Errorf("failed to load auth data: %w", err)
	}
	if authData == nil {
		return errors.New("no pending authentication request")
	}

	// Sign up with the provided information
	_, err = c.api.AuthSignUp(c.ctx, &tg.AuthSignUpRequest{
		PhoneNumber:   authData.PhoneNumber,
		PhoneCodeHash: authData.PhoneCodeHash,
		FirstName:     firstName,
		LastName:      lastName,
	})

	if err != nil {
		c.logger.Error("Failed to register", "error", err)
		return fmt.Errorf("failed to register: %w", err)
	}

	// Registration successful
	c.setAuthState(types.AuthStateReady)

	// Get current user info
	user, err := c.api.UsersGetFullUser(c.ctx, &tg.InputUserSelf{})
	if err == nil {
		if u, ok := user.Users[0].(*tg.User); ok {
			c.authMu.Lock()
			c.currentUser = u
			c.authMu.Unlock()
		}
	}

	// Clear temporary auth data
	c.sessionStorage.ClearAuthData()

	c.logger.Info("Registration successful")
	return nil
}

// Logout logs out from the current session.
func (c *Client) Logout() error {
	c.logger.Info("Logging out")

	// Call logout API
	_, err := c.api.AuthLogOut(c.ctx)
	if err != nil {
		c.logger.Error("Failed to logout", "error", err)
		return fmt.Errorf("failed to logout: %w", err)
	}

	// Clear session data
	c.setAuthState(types.AuthStateWaitPhoneNumber)
	c.authMu.Lock()
	c.currentUser = nil
	c.authMu.Unlock()

	c.logger.Info("Logout successful")
	return nil
}

// Implementation of auth.FlowClient interface for authFlow

// Phone returns the phone number for authentication.
func (f *authFlow) Phone(_ context.Context) (string, error) {
	return f.phoneNumber, nil
}

// Password returns the password for 2FA (not used in this implementation).
func (f *authFlow) Password(_ context.Context) (string, error) {
	// We handle password separately via SendPassword
	return "", errors.New("password required, use SendPassword")
}

// Code returns the authentication code (not used in this implementation).
func (f *authFlow) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	// We handle code separately via SendAuthCode
	return "", errors.New("code required, use SendAuthCode")
}

// AcceptTermsOfService accepts Telegram's terms of service.
func (f *authFlow) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	f.client.logger.Info("Accepting terms of service", "tos_id", tos.ID)
	return nil
}

// SignUp provides registration information (not used in this implementation).
func (f *authFlow) SignUp(_ context.Context) (auth.UserInfo, error) {
	// We handle registration separately via SendRegistration
	return auth.UserInfo{}, errors.New("registration required, use SendRegistration")
}
