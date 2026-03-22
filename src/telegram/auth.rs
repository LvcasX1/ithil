//! Authentication methods for the Telegram client.
//!
//! This module provides the authentication flow implementation:
//! 1. Request login code via phone number
//! 2. Sign in with the received code
//! 3. Provide 2FA password if required
//! 4. Sign up if the phone number is not registered

use grammers_client::SignInError;
use tracing::{debug, info, warn};

use super::client::TelegramClient;
use super::error::TelegramError;
use crate::types::{AuthState, User, UserStatus};

impl TelegramClient {
    /// Requests a login code for the given phone number.
    ///
    /// The code will be sent via SMS or another Telegram app. After calling
    /// this method, the auth state will change to [`AuthState::WaitCode`].
    ///
    /// # Arguments
    ///
    /// * `phone` - Phone number in international format (e.g., "+1234567890")
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected
    /// - The phone number is invalid
    /// - Network error occurs
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// client.request_login_code("+1234567890").await?;
    /// // Now ask the user for the code they received
    /// # Ok(())
    /// # }
    /// ```
    pub async fn request_login_code(&self, phone: &str) -> Result<(), TelegramError> {
        let client = self.client().await?;

        info!(
            "Requesting login code for phone: {}***",
            &phone[..4.min(phone.len())]
        );

        // request_login_code takes phone and api_hash
        let token = client
            .request_login_code(phone, self.api_hash())
            .await
            .map_err(TelegramError::from)?;

        // Store the token for use in sign_in
        self.set_login_token(token).await;
        self.set_auth_state(AuthState::WaitCode).await;

        debug!("Login code requested successfully");
        Ok(())
    }

    /// Signs in with the verification code received via SMS/Telegram.
    ///
    /// Call this after [`request_login_code`](Self::request_login_code). The auth state
    /// will change to [`AuthState::Ready`] on success.
    ///
    /// # Arguments
    ///
    /// * `code` - The verification code received (e.g., "12345")
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected
    /// - No login code was requested (call `request_login_code` first)
    /// - The code is invalid
    /// - 2FA password is required (auth state changes to `WaitPassword`)
    /// - Sign up is required (auth state changes to `WaitRegistration`)
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// // After requesting login code and getting the code from the user
    /// match client.sign_in("12345").await {
    ///     Ok(()) => println!("Signed in successfully!"),
    ///     Err(e) => println!("Sign in failed: {}", e),
    /// }
    /// # Ok(())
    /// # }
    /// ```
    pub async fn sign_in(&self, code: &str) -> Result<(), TelegramError> {
        let client = self.client().await?;

        let token = self
            .take_login_token()
            .await
            .ok_or_else(|| TelegramError::Internal("No login token available".into()))?;

        info!("Attempting to sign in with verification code");

        match client.sign_in(&token, code).await {
            Ok(user) => {
                let name = user.first_name().unwrap_or("Unknown");
                info!("Signed in successfully as: {}", name);
                self.set_auth_state(AuthState::Ready).await;

                Ok(())
            },
            Err(SignInError::PasswordRequired(password_token)) => {
                info!("2FA password required");
                self.set_password_token(password_token).await;
                self.set_auth_state(AuthState::WaitPassword).await;
                Err(TelegramError::PasswordRequired)
            },
            Err(SignInError::SignUpRequired) => {
                info!("Sign up required - phone number not registered");
                // Store the login token back for sign_up
                self.set_login_token(token).await;
                self.set_auth_state(AuthState::WaitRegistration).await;
                Err(TelegramError::SignUpRequired)
            },
            Err(SignInError::InvalidCode) => {
                warn!("Invalid verification code");
                // Store the token back so user can retry
                self.set_login_token(token).await;
                Err(TelegramError::InvalidCode)
            },
            Err(e) => {
                warn!("Sign in failed: {}", e);
                Err(e.into())
            },
        }
    }

    /// Provides the 2FA password when required.
    ///
    /// Call this after [`sign_in`](Self::sign_in) returns
    /// [`TelegramError::PasswordRequired`].
    ///
    /// # Arguments
    ///
    /// * `password` - The user's 2FA password
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected
    /// - No password token is available (`sign_in` didn't request password)
    /// - The password is incorrect
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # use ithil::telegram::TelegramError;
    /// # async fn example(client: &TelegramClient) -> Result<(), TelegramError> {
    /// // After sign_in returns PasswordRequired
    /// client.check_password("my_secure_password").await?;
    /// # Ok(())
    /// # }
    /// ```
    pub async fn check_password(&self, password: &str) -> Result<(), TelegramError> {
        let client = self.client().await?;

        let token = self
            .take_password_token()
            .await
            .ok_or_else(|| TelegramError::Internal("No password token available".into()))?;

        info!("Checking 2FA password");

        match client.check_password(token, password).await {
            Ok(user) => {
                let name = user.first_name().unwrap_or("Unknown");
                info!("Signed in with 2FA as: {}", name);
                self.set_auth_state(AuthState::Ready).await;

                Ok(())
            },
            Err(SignInError::InvalidPassword(_new_token)) => {
                warn!("Invalid 2FA password");
                // User needs to get a new code and start over
                self.set_auth_state(AuthState::WaitPhoneNumber).await;
                Err(TelegramError::InvalidPassword)
            },
            Err(e) => {
                warn!("2FA check failed: {}", e);
                self.set_auth_state(AuthState::WaitPhoneNumber).await;
                Err(e.into())
            },
        }
    }

    /// Signs up as a new user.
    ///
    /// Call this after [`sign_in`](Self::sign_in) returns
    /// [`TelegramError::SignUpRequired`].
    ///
    /// Note: Sign up via the API is restricted. Users should typically
    /// sign up via an official Telegram app first.
    ///
    /// # Arguments
    ///
    /// * `first_name` - User's first name
    /// * `last_name` - User's last name (can be empty)
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - The client is not connected
    /// - No login token is available
    /// - Sign up fails (API restrictions, etc.)
    #[allow(clippy::unused_async)]
    pub async fn sign_up(&self, _first_name: &str, _last_name: &str) -> Result<(), TelegramError> {
        // Note: Sign up via third-party apps is not supported by Telegram anymore.
        // Users must sign up using an official Telegram app first.
        // See: https://bugs.telegram.org/c/25410/1
        warn!(
            "Sign up via third-party apps is not supported. Please use an official Telegram app."
        );
        Err(TelegramError::SignUpRequired)
    }

    /// Gets information about the currently logged-in user.
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected or not authorized.
    ///
    /// # Examples
    ///
    /// ```rust,no_run
    /// # use ithil::telegram::TelegramClient;
    /// # async fn example(client: &TelegramClient) -> Result<(), ithil::telegram::TelegramError> {
    /// let me = client.get_me().await?;
    /// println!("Logged in as: {}", me.get_display_name());
    /// # Ok(())
    /// # }
    /// ```
    pub async fn get_me(&self) -> Result<User, TelegramError> {
        let client = self.require_authorized().await?;

        let me = client.get_me().await.map_err(TelegramError::from)?;

        Ok(grammers_user_to_user(&me))
    }

    /// Logs out and clears the session.
    ///
    /// After calling this, the user will need to re-authenticate.
    ///
    /// # Errors
    ///
    /// Returns an error if the client is not connected.
    pub async fn log_out(&self) -> Result<(), TelegramError> {
        let client = self.client().await?;

        info!("Logging out...");

        client.sign_out().await.map_err(TelegramError::from)?;

        // Clear the session file
        let session_path = self.session_path();
        if std::path::Path::new(session_path).exists() {
            if let Err(e) = std::fs::remove_file(session_path) {
                warn!("Failed to remove session file: {}", e);
            }
        }

        self.set_auth_state(AuthState::WaitPhoneNumber).await;
        info!("Logged out successfully");

        Ok(())
    }
}

/// Converts a grammers User to our User type.
pub(crate) fn grammers_user_to_user(user: &grammers_client::peer::User) -> User {
    User {
        id: user.id().bare_id(),
        first_name: user.first_name().unwrap_or("").to_string(),
        last_name: user.last_name().unwrap_or("").to_string(),
        username: user.username().unwrap_or("").to_string(),
        phone_number: user.phone().unwrap_or("").to_string(),
        profile_photo_id: String::new(), // Photo handling would require additional API calls
        status: UserStatus::Offline,     // Would need to check user.raw for status
        is_bot: user.is_bot(),
        is_contact: false, // Not directly available
        is_mutual_contact: false,
        is_verified: user.verified(),
        is_premium: user.is_premium(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_grammers_user_conversion_defaults() {
        // We can't easily create a grammers User for testing without mocking,
        // but we can at least verify our function signature is correct
        // and that the User struct has all expected fields
        let user = User::default();
        assert_eq!(user.id, 0);
        assert!(user.first_name.is_empty());
        assert!(!user.is_bot);
    }
}
