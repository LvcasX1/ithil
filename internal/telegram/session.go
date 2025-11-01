// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
)

// SessionStorage implements persistent session storage to disk.
type SessionStorage struct {
	path string
}

// NewSessionStorage creates a new session storage instance.
func NewSessionStorage(path string) *SessionStorage {
	return &SessionStorage{path: path}
}

// LoadSession loads session data from disk.
func (s *SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	// Ensure the directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		// No session file exists yet, return empty
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// StoreSession stores session data to disk.
func (s *SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	// Ensure the directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Write to a temporary file first, then rename (atomic operation)
	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}


// AuthData stores authentication state information.
type AuthData struct {
	PhoneNumber  string `json:"phone_number"`
	PhoneCodeHash string `json:"phone_code_hash"`
	IsRegistered bool   `json:"is_registered"`
}

// SaveAuthData saves authentication data to a separate file.
func (s *SessionStorage) SaveAuthData(data *AuthData) error {
	authPath := s.path + ".auth"

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dir := filepath.Dir(authPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(authPath, jsonData, 0600)
}

// LoadAuthData loads authentication data from disk.
func (s *SessionStorage) LoadAuthData() (*AuthData, error) {
	authPath := s.path + ".auth"

	data, err := os.ReadFile(authPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var authData AuthData
	if err := json.Unmarshal(data, &authData); err != nil {
		return nil, err
	}

	return &authData, nil
}

// ClearAuthData removes the authentication data file.
func (s *SessionStorage) ClearAuthData() error {
	authPath := s.path + ".auth"
	err := os.Remove(authPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ClearSession removes the session file (for invalid/corrupted sessions).
func (s *SessionStorage) ClearSession() error {
	err := os.Remove(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
