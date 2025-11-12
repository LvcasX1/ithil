//go:build cgo

// Package media provides media playback utilities.
package media

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

// PlaybackState represents the current state of audio playback.
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
)

// AudioPlayer handles audio file playback using the Beep library.
type AudioPlayer struct {
	mu              sync.RWMutex
	state           PlaybackState
	streamer        beep.StreamSeekCloser
	ctrl            *beep.Ctrl
	volume          *effects.Volume
	format          beep.Format
	filePath        string
	duration        time.Duration
	speakerInit     bool
	speakerInitOnce sync.Once
	position        time.Duration
	ticker          *time.Ticker
	stopTicker      chan bool
}

// NewAudioPlayer creates a new audio player instance.
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		state:      StateStopped,
		stopTicker: make(chan bool),
	}
}

// LoadFile loads an audio file for playback.
// Supports: .mp3, .ogg, .wav formats
func (p *AudioPlayer) LoadFile(filePath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any existing playback
	if p.state != StateStopped {
		p.stopPlayback()
	}

	// Open the audio file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}

	// Decode based on file extension
	ext := filepath.Ext(filePath)
	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to decode MP3: %w", err)
		}

	case ".ogg":
		streamer, format, err = vorbis.Decode(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to decode OGG/Vorbis: %w", err)
		}

	case ".wav":
		streamer, format, err = wav.Decode(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to decode WAV: %w", err)
		}

	default:
		file.Close()
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Initialize speaker if needed (only once globally)
	initErr := p.initSpeaker(format.SampleRate)
	if initErr != nil {
		streamer.Close()
		return fmt.Errorf("failed to initialize audio device: %w", initErr)
	}

	// Calculate duration
	duration := format.SampleRate.D(streamer.Len())

	// Create control wrapper for pause/resume
	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}

	// Create volume control (default: 0 = 100%)
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}

	// Store player state
	p.streamer = streamer
	p.ctrl = ctrl
	p.volume = volume
	p.format = format
	p.filePath = filePath
	p.duration = duration
	p.position = 0

	return nil
}

// Play starts or resumes audio playback.
func (p *AudioPlayer) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return fmt.Errorf("no audio file loaded")
	}

	switch p.state {
	case StatePlaying:
		// Already playing, do nothing
		return nil

	case StatePaused:
		// Resume playback
		speaker.Lock()
		p.ctrl.Paused = false
		speaker.Unlock()
		p.state = StatePlaying
		p.startPositionTracking()

	case StateStopped:
		// Start new playback
		speaker.Play(beep.Seq(p.volume, beep.Callback(func() {
			p.mu.Lock()
			p.state = StateStopped
			p.position = p.duration
			p.stopPositionTracking()
			p.mu.Unlock()
		})))
		p.state = StatePlaying
		p.startPositionTracking()
	}

	return nil
}

// Pause pauses audio playback.
func (p *AudioPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return fmt.Errorf("no audio file loaded")
	}

	if p.state == StatePlaying {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
		p.state = StatePaused
		p.stopPositionTracking()
	}

	return nil
}

// Stop stops audio playback and resets position.
func (p *AudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return nil
	}

	p.stopPlayback()
	return nil
}

// stopPlayback internal stop logic (must be called with lock held)
func (p *AudioPlayer) stopPlayback() {
	if p.ctrl != nil {
		speaker.Clear()
		p.stopPositionTracking()
	}

	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}

	p.ctrl = nil
	p.volume = nil
	p.state = StateStopped
	p.position = 0
}

// TogglePlayPause toggles between play and pause states.
func (p *AudioPlayer) TogglePlayPause() error {
	p.mu.RLock()
	state := p.state
	p.mu.RUnlock()

	if state == StatePlaying {
		return p.Pause()
	}
	return p.Play()
}

// Seek seeks to a specific position in the audio file.
func (p *AudioPlayer) Seek(position time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return fmt.Errorf("no audio file loaded")
	}

	// Clamp position to valid range
	if position < 0 {
		position = 0
	}
	if position > p.duration {
		position = p.duration
	}

	// Calculate sample position
	samplePos := p.format.SampleRate.N(position)

	// Seek in the streamer
	speaker.Lock()
	err := p.streamer.Seek(samplePos)
	speaker.Unlock()

	if err != nil {
		return fmt.Errorf("seek failed: %w", err)
	}

	p.position = position
	return nil
}

// SkipForward skips forward by the specified duration.
func (p *AudioPlayer) SkipForward(delta time.Duration) error {
	p.mu.RLock()
	newPos := p.position + delta
	p.mu.RUnlock()

	return p.Seek(newPos)
}

// SkipBackward skips backward by the specified duration.
func (p *AudioPlayer) SkipBackward(delta time.Duration) error {
	p.mu.RLock()
	newPos := p.position - delta
	p.mu.RUnlock()

	return p.Seek(newPos)
}

// SetVolume sets the playback volume.
// volume: -5 (silent) to 0 (100%) to 5 (amplified)
func (p *AudioPlayer) SetVolume(volume float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volume == nil {
		return fmt.Errorf("no audio file loaded")
	}

	// Clamp volume to reasonable range
	if volume < -5 {
		volume = -5
	}
	if volume > 5 {
		volume = 5
	}

	speaker.Lock()
	p.volume.Volume = volume
	speaker.Unlock()

	return nil
}

// VolumeUp increases volume by 0.5
func (p *AudioPlayer) VolumeUp() error {
	p.mu.RLock()
	if p.volume == nil {
		p.mu.RUnlock()
		return fmt.Errorf("no audio file loaded")
	}
	currentVol := p.volume.Volume
	p.mu.RUnlock()

	return p.SetVolume(currentVol + 0.5)
}

// VolumeDown decreases volume by 0.5
func (p *AudioPlayer) VolumeDown() error {
	p.mu.RLock()
	if p.volume == nil {
		p.mu.RUnlock()
		return fmt.Errorf("no audio file loaded")
	}
	currentVol := p.volume.Volume
	p.mu.RUnlock()

	return p.SetVolume(currentVol - 0.5)
}

// GetState returns the current playback state.
func (p *AudioPlayer) GetState() PlaybackState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetPosition returns the current playback position.
func (p *AudioPlayer) GetPosition() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.position
}

// GetDuration returns the total duration of the loaded audio file.
func (p *AudioPlayer) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.duration
}

// GetVolume returns the current volume level.
func (p *AudioPlayer) GetVolume() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.volume == nil {
		return 0
	}
	return p.volume.Volume
}

// GetFilePath returns the path of the currently loaded file.
func (p *AudioPlayer) GetFilePath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.filePath
}

// Close cleans up resources used by the audio player.
func (p *AudioPlayer) Close() error {
	return p.Stop()
}

// initSpeaker initializes the speaker (only once globally).
func (p *AudioPlayer) initSpeaker(sampleRate beep.SampleRate) error {
	var initErr error
	p.speakerInitOnce.Do(func() {
		// Initialize speaker with buffer size of 1/10 second
		err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
		if err != nil {
			initErr = err
			return
		}
		p.speakerInit = true
	})
	return initErr
}

// startPositionTracking starts tracking playback position.
func (p *AudioPlayer) startPositionTracking() {
	// Stop any existing ticker
	p.stopPositionTracking()

	// Start new ticker
	p.ticker = time.NewTicker(100 * time.Millisecond)
	p.stopTicker = make(chan bool)

	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.mu.Lock()
				if p.streamer != nil && p.state == StatePlaying {
					speaker.Lock()
					currentSample := p.format.SampleRate.D(p.streamer.Position())
					speaker.Unlock()
					p.position = currentSample
				}
				p.mu.Unlock()
			case <-p.stopTicker:
				return
			}
		}
	}()
}

// stopPositionTracking stops tracking playback position.
func (p *AudioPlayer) stopPositionTracking() {
	if p.ticker != nil {
		p.ticker.Stop()
		close(p.stopTicker)
		p.ticker = nil
	}
}
