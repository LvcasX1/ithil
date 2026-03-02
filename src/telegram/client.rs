//! Main Telegram client wrapper.
//!
//! This module provides the [`TelegramClient`] struct which wraps grammers
//! to provide a high-level interface for Telegram operations.

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

use grammers_client::client::{LoginToken, PasswordToken, UpdatesConfiguration, UpdateStream};
use grammers_client::sender::SenderPoolFatHandle;
use grammers_client::{Client, SenderPool};
use grammers_session::storages::SqliteSession;
use grammers_session::updates::UpdatesLike;
use tokio::sync::{mpsc, RwLock};
use tokio::task::JoinHandle;
use tracing::{debug, info};

use super::error::TelegramError;
use crate::cache::SharedCache;
use crate::types::{AuthState, Update};

/// Main Telegram client wrapper.
///
/// This struct provides a high-level interface for all Telegram operations,
/// managing the connection lifecycle, authentication, and real-time updates.
///
/// # Thread Safety
///
/// `TelegramClient` is designed to be shared across threads using `Arc`.
/// Internal state is protected by `RwLock` for concurrent access.
///
/// # Session Management
///
/// The session is automatically loaded from disk on connect and saved on disconnect.
/// The session file stores authentication state so users don't need to re-authenticate
/// on every app restart.
pub struct TelegramClient {
    /// The underlying grammers client (None when disconnected)
    inner: Arc<RwLock<Option<Client>>>,

    /// Telegram API ID
    api_id: i32,

    /// Telegram API hash
    api_hash: String,

    /// Path to the session file
    session_path: String,

    /// Current authentication state
    auth_state: Arc<RwLock<AuthState>>,

    /// Channel for sending updates to the UI
    update_tx: Arc<RwLock<Option<mpsc::Sender<Update>>>>,

    /// Shared cache for storing Telegram data
    cache: SharedCache,

    /// Flag indicating if the update loop is running
    update_loop_running: Arc<AtomicBool>,

    /// Login token stored between request_login_code and sign_in
    login_token: Arc<RwLock<Option<LoginToken>>>,

    /// Password token stored for 2FA authentication
    password_token: Arc<RwLock<Option<PasswordToken>>>,

    /// Session storage for saving
    session: Arc<RwLock<Option<Arc<SqliteSession>>>>,

    /// Handle to the sender pool runner task
    pool_task: Arc<RwLock<Option<JoinHandle<()>>>>,

    /// Handle to quit the sender pool
    pool_handle: Arc<RwLock<Option<SenderPoolFatHandle>>>,

    /// Receiver for raw updates from the sender pool
    updates_receiver: Arc<RwLock<Option<mpsc::UnboundedReceiver<UpdatesLike>>>>,
}

impl TelegramClient {
    /// Creates a new Telegram client instance.
    ///
    /// This does not connect to Telegram - call [`connect`](Self::connect) to establish
    /// a connection.
    ///
    /// # Arguments
    ///
    /// * `api_id` - Telegram API ID obtained from <https://my.telegram.org>
    /// * `api_hash` - Telegram API hash obtained from <https://my.telegram.org>
    /// * `session_path` - Path to the session file for storing authentication state
    /// * `cache` - Shared cache for storing chats, messages, and users
    ///
    /// # Examples
    ///
    /// ```rust
    /// use ithil::telegram::TelegramClient;
    /// use ithil::cache::new_shared_cache;
    ///
    /// let cache = new_shared_cache(100);
    /// let client = TelegramClient::new(
    ///     12345,
    ///     "api_hash".to_string(),
    ///     "session.session".to_string(),
    ///     cache,
    /// );
    /// ```
    #[must_use]
    pub fn new(
        api_id: i32,
        api_hash: String,
        session_path: String,
        cache: SharedCache,
    ) -> Self {
        Self {
            inner: Arc::new(RwLock::new(None)),
            api_id,
            api_hash,
            session_path,
            auth_state: Arc::new(RwLock::new(AuthState::WaitPhoneNumber)),
            update_tx: Arc::new(RwLock::new(None)),
            cache,
            update_loop_running: Arc::new(AtomicBool::new(false)),
            login_token: Arc::new(RwLock::new(None)),
            password_token: Arc::new(RwLock::new(None)),
            session: Arc::new(RwLock::new(None)),
            pool_task: Arc::new(RwLock::new(None)),
            pool_handle: Arc::new(RwLock::new(None)),
            updates_receiver: Arc::new(RwLock::new(None)),
        }
    }

    /// Connects to Telegram servers.
    ///
    /// This method will:
    /// 1. Load or create a session file
    /// 2. Establish a connection to Telegram
    /// 3. Check if the user is already authorized
    ///
    /// After connecting, check [`get_auth_state`](Self::get_auth_state) to determine
    /// if the user needs to authenticate.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The session file cannot be loaded or created
    /// - The connection to Telegram fails
    pub async fn connect(&self) -> Result<(), TelegramError> {
        info!("Connecting to Telegram...");

        // Load or create session using SqliteSession
        let session = Arc::new(
            SqliteSession::open(&self.session_path)
                .await
                .map_err(|e| TelegramError::Session(e.to_string()))?
        );

        debug!("Session loaded from {}", self.session_path);

        // Create SenderPool and Client
        let SenderPool {
            runner,
            handle,
            updates,
        } = SenderPool::new(Arc::clone(&session), self.api_id);

        let client = Client::new(handle.clone());

        // Spawn the runner task
        let pool_task = tokio::spawn(runner.run());

        info!("Connected to Telegram servers");

        // Check if already authorized
        let is_authorized = client
            .is_authorized()
            .await
            .map_err(TelegramError::from)?;

        // Update auth state
        let new_state = if is_authorized {
            info!("User is already authorized");
            AuthState::Ready
        } else {
            debug!("User needs to authenticate");
            AuthState::WaitPhoneNumber
        };

        *self.auth_state.write().await = new_state;
        *self.inner.write().await = Some(client);
        *self.session.write().await = Some(session);
        *self.pool_task.write().await = Some(pool_task);
        *self.pool_handle.write().await = Some(handle);
        *self.updates_receiver.write().await = Some(updates);

        Ok(())
    }

    /// Disconnects from Telegram and saves the session.
    ///
    /// This method will:
    /// 1. Stop the update loop if running
    /// 2. Save the session to disk
    /// 3. Close the connection
    ///
    /// # Errors
    ///
    /// Returns an error if the session cannot be saved.
    pub async fn disconnect(&self) -> Result<(), TelegramError> {
        info!("Disconnecting from Telegram...");

        // Stop update loop
        self.stop_update_loop().await;

        // Quit the pool handle to signal disconnection
        if let Some(handle) = self.pool_handle.write().await.take() {
            handle.quit();
        }

        // Wait for the pool task to finish
        if let Some(task) = self.pool_task.write().await.take() {
            let _ = task.await;
        }

        // Clear the client and session
        *self.inner.write().await = None;
        *self.session.write().await = None;

        *self.auth_state.write().await = AuthState::Closed;
        info!("Disconnected from Telegram");

        Ok(())
    }

    /// Returns `true` if the client is connected to Telegram.
    ///
    /// Note: This checks if the client object exists, not if the network
    /// connection is still active. Use [`get_auth_state`](Self::get_auth_state)
    /// for a more complete picture of the connection status.
    pub async fn is_connected(&self) -> bool {
        self.inner.read().await.is_some()
    }

    /// Gets the current authentication state.
    ///
    /// Use this to determine what action the user needs to take:
    ///
    /// - [`AuthState::WaitPhoneNumber`] - User needs to enter phone number
    /// - [`AuthState::WaitCode`] - User needs to enter verification code
    /// - [`AuthState::WaitPassword`] - User needs to enter 2FA password
    /// - [`AuthState::WaitRegistration`] - User needs to register (new account)
    /// - [`AuthState::Ready`] - User is authenticated and ready
    /// - [`AuthState::Closed`] - Connection is closed
    pub async fn get_auth_state(&self) -> AuthState {
        *self.auth_state.read().await
    }

    /// Sets the channel for streaming updates to the UI.
    ///
    /// Updates from Telegram (new messages, chat updates, etc.) will be sent
    /// through this channel. The UI should poll this channel to receive updates.
    ///
    /// # Arguments
    ///
    /// * `tx` - The sender half of an mpsc channel for `Update` messages
    pub async fn set_update_channel(&self, tx: mpsc::Sender<Update>) {
        *self.update_tx.write().await = Some(tx);
    }

    /// Gets a clone of the underlying grammers client.
    ///
    /// # Errors
    ///
    /// Returns [`TelegramError::NotConnected`] if the client is not connected.
    pub(crate) async fn client(&self) -> Result<Client, TelegramError> {
        self.inner
            .read()
            .await
            .clone()
            .ok_or(TelegramError::NotConnected)
    }

    /// Ensures the client is connected and authorized.
    ///
    /// # Errors
    ///
    /// Returns an error if not connected or not authorized.
    pub(crate) async fn require_authorized(&self) -> Result<Client, TelegramError> {
        let client = self.client().await?;
        let state = self.get_auth_state().await;

        if state != AuthState::Ready {
            return Err(TelegramError::AuthRequired);
        }

        Ok(client)
    }

    /// Gets the shared cache.
    #[must_use]
    pub fn cache(&self) -> &SharedCache {
        &self.cache
    }

    /// Gets the API hash (needed for auth methods).
    pub(crate) fn api_hash(&self) -> &str {
        &self.api_hash
    }

    /// Gets the session path.
    pub(crate) fn session_path(&self) -> &str {
        &self.session_path
    }

    /// Returns `true` if the update loop is currently running.
    pub fn is_update_loop_running(&self) -> bool {
        self.update_loop_running.load(Ordering::SeqCst)
    }

    /// Internal: Updates the authentication state.
    pub(crate) async fn set_auth_state(&self, state: AuthState) {
        let mut guard = self.auth_state.write().await;
        if *guard != state {
            info!("Auth state changed: {:?} -> {:?}", *guard, state);
            *guard = state;
        }
    }

    /// Internal: Stores the login token for sign_in.
    pub(crate) async fn set_login_token(&self, token: LoginToken) {
        *self.login_token.write().await = Some(token);
    }

    /// Internal: Gets and clears the stored login token.
    pub(crate) async fn take_login_token(&self) -> Option<LoginToken> {
        self.login_token.write().await.take()
    }

    /// Internal: Stores the password token for 2FA.
    pub(crate) async fn set_password_token(&self, token: PasswordToken) {
        *self.password_token.write().await = Some(token);
    }

    /// Internal: Gets and clears the stored password token.
    pub(crate) async fn take_password_token(&self) -> Option<PasswordToken> {
        self.password_token.write().await.take()
    }

    /// Internal: Gets the update sender channel.
    pub(crate) async fn get_update_sender(&self) -> Option<mpsc::Sender<Update>> {
        self.update_tx.read().await.clone()
    }

    /// Internal: Sets the update loop running flag.
    pub(crate) fn set_update_loop_running(&self, running: bool) {
        self.update_loop_running.store(running, Ordering::SeqCst);
    }

    /// Internal: Takes the updates receiver to create an update stream.
    ///
    /// Returns `None` if already taken or if not connected.
    pub(crate) async fn take_updates_receiver(
        &self,
    ) -> Option<tokio::sync::mpsc::UnboundedReceiver<UpdatesLike>> {
        self.updates_receiver.write().await.take()
    }

    /// Internal: Creates an update stream for receiving real-time updates.
    ///
    /// This takes ownership of the updates receiver, so it can only be called once
    /// per connection.
    pub(crate) async fn create_update_stream(&self) -> Result<UpdateStream, TelegramError> {
        let client = self.require_authorized().await?;
        let updates = self
            .take_updates_receiver()
            .await
            .ok_or_else(|| TelegramError::Internal("Updates receiver already taken".into()))?;

        let config = UpdatesConfiguration::default();
        Ok(client.stream_updates(updates, config).await)
    }
}

impl std::fmt::Debug for TelegramClient {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("TelegramClient")
            .field("api_id", &self.api_id)
            .field("session_path", &self.session_path)
            .field("update_loop_running", &self.update_loop_running)
            .finish_non_exhaustive()
    }
}

// TelegramClient is cloneable for spawning tasks - all fields use Arc
impl Clone for TelegramClient {
    fn clone(&self) -> Self {
        Self {
            inner: Arc::clone(&self.inner),
            api_id: self.api_id,
            api_hash: self.api_hash.clone(),
            session_path: self.session_path.clone(),
            auth_state: Arc::clone(&self.auth_state),
            update_tx: Arc::clone(&self.update_tx),
            cache: self.cache.clone(),
            update_loop_running: Arc::clone(&self.update_loop_running),
            login_token: Arc::clone(&self.login_token),
            password_token: Arc::clone(&self.password_token),
            session: Arc::clone(&self.session),
            pool_task: Arc::clone(&self.pool_task),
            pool_handle: Arc::clone(&self.pool_handle),
            updates_receiver: Arc::clone(&self.updates_receiver),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::cache::new_shared_cache;

    #[test]
    fn test_new_client() {
        let cache = new_shared_cache(100);
        let client = TelegramClient::new(
            12345,
            "test_hash".to_string(),
            "test.session".to_string(),
            cache,
        );

        assert_eq!(client.api_id, 12345);
        assert_eq!(client.api_hash, "test_hash");
        assert_eq!(client.session_path, "test.session");
        assert!(!client.is_update_loop_running());
    }

    #[tokio::test]
    async fn test_initial_state() {
        let cache = new_shared_cache(100);
        let client = TelegramClient::new(
            12345,
            "test_hash".to_string(),
            "test.session".to_string(),
            cache,
        );

        assert!(!client.is_connected().await);
        assert_eq!(client.get_auth_state().await, AuthState::WaitPhoneNumber);
    }

    #[test]
    fn test_debug_impl() {
        let cache = new_shared_cache(100);
        let client = TelegramClient::new(
            12345,
            "test_hash".to_string(),
            "test.session".to_string(),
            cache,
        );

        let debug_str = format!("{:?}", client);
        assert!(debug_str.contains("TelegramClient"));
        assert!(debug_str.contains("12345"));
        assert!(debug_str.contains("test.session"));
        // API hash should not be in debug output for security
        assert!(!debug_str.contains("test_hash"));
    }
}
