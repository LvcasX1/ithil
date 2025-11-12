//go:build !cgo

// Package media provides media playback utilities.
package media

import (
	"fmt"
	"sync"
	"time"
)

// PlaybackState represents the current state of audio playback.
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
)

// AudioPlayer handles audio file playback (stub implementation without CGO).
type AudioPlayer struct {
	mu sync.RWMutex
}

// NewAudioPlayer creates a new audio player instance.
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{}
}

// LoadFile returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) LoadFile(filePath string) error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// Play returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) Play() error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// Pause returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) Pause() error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// Stop returns nil (no-op).
func (p *AudioPlayer) Stop() error {
	return nil
}

// TogglePlayPause returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) TogglePlayPause() error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// Seek returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) Seek(position time.Duration) error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// SkipForward returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) SkipForward(delta time.Duration) error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// SkipBackward returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) SkipBackward(delta time.Duration) error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// SetVolume returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) SetVolume(volume float64) error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// VolumeUp returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) VolumeUp() error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// VolumeDown returns an error indicating audio is not supported without CGO.
func (p *AudioPlayer) VolumeDown() error {
	return fmt.Errorf("audio playback not supported in this build (requires CGO)")
}

// GetState returns StateStopped.
func (p *AudioPlayer) GetState() PlaybackState {
	return StateStopped
}

// GetPosition returns 0.
func (p *AudioPlayer) GetPosition() time.Duration {
	return 0
}

// GetDuration returns 0.
func (p *AudioPlayer) GetDuration() time.Duration {
	return 0
}

// GetVolume returns 0.
func (p *AudioPlayer) GetVolume() float64 {
	return 0
}

// GetFilePath returns an empty string.
func (p *AudioPlayer) GetFilePath() string {
	return ""
}

// Close returns nil (no-op).
func (p *AudioPlayer) Close() error {
	return nil
}
