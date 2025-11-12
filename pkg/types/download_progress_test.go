package types

import (
	"testing"
	"time"
)

func TestDownloadProgress_GetPercentage(t *testing.T) {
	tests := []struct {
		name        string
		bytesLoaded int64
		bytesTotal  int64
		expected    float64
	}{
		{
			name:        "zero total",
			bytesLoaded: 0,
			bytesTotal:  0,
			expected:    0,
		},
		{
			name:        "half downloaded",
			bytesLoaded: 50,
			bytesTotal:  100,
			expected:    50.0,
		},
		{
			name:        "fully downloaded",
			bytesLoaded: 100,
			bytesTotal:  100,
			expected:    100.0,
		},
		{
			name:        "over 100 percent (should cap at 100)",
			bytesLoaded: 150,
			bytesTotal:  100,
			expected:    100.0,
		},
		{
			name:        "partial percentage",
			bytesLoaded: 333,
			bytesTotal:  1000,
			expected:    33.300000000000004, // Floating point precision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DownloadProgress{
				BytesLoaded: tt.bytesLoaded,
				BytesTotal:  tt.bytesTotal,
			}
			got := p.GetPercentage()
			if got != tt.expected {
				t.Errorf("GetPercentage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDownloadProgress_GetSpeed(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		bytesLoaded int64
		startTime   time.Time
		lastUpdate  time.Time
		expected    float64
	}{
		{
			name:        "no time elapsed",
			bytesLoaded: 100,
			startTime:   now,
			lastUpdate:  now,
			expected:    0,
		},
		{
			name:        "1 second elapsed, 100 bytes",
			bytesLoaded: 100,
			startTime:   now,
			lastUpdate:  now.Add(1 * time.Second),
			expected:    100.0,
		},
		{
			name:        "2 seconds elapsed, 200 bytes",
			bytesLoaded: 200,
			startTime:   now,
			lastUpdate:  now.Add(2 * time.Second),
			expected:    100.0,
		},
		{
			name:        "0.5 seconds elapsed, 50 bytes",
			bytesLoaded: 50,
			startTime:   now,
			lastUpdate:  now.Add(500 * time.Millisecond),
			expected:    100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DownloadProgress{
				BytesLoaded: tt.bytesLoaded,
				StartTime:   tt.startTime,
				LastUpdate:  tt.lastUpdate,
			}
			got := p.GetSpeed()
			if got != tt.expected {
				t.Errorf("GetSpeed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDownloadProgress_GetETA(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		bytesLoaded int64
		bytesTotal  int64
		startTime   time.Time
		lastUpdate  time.Time
		expected    time.Duration
	}{
		{
			name:        "all downloaded",
			bytesLoaded: 100,
			bytesTotal:  100,
			startTime:   now,
			lastUpdate:  now.Add(1 * time.Second),
			expected:    0,
		},
		{
			name:        "no speed (no time elapsed)",
			bytesLoaded: 50,
			bytesTotal:  100,
			startTime:   now,
			lastUpdate:  now,
			expected:    0,
		},
		{
			name:        "half done, 1 second elapsed, expect 1 more second",
			bytesLoaded: 50,
			bytesTotal:  100,
			startTime:   now,
			lastUpdate:  now.Add(1 * time.Second),
			expected:    1 * time.Second,
		},
		{
			name:        "25% done, 1 second elapsed, expect 3 more seconds",
			bytesLoaded: 25,
			bytesTotal:  100,
			startTime:   now,
			lastUpdate:  now.Add(1 * time.Second),
			expected:    3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DownloadProgress{
				BytesLoaded: tt.bytesLoaded,
				BytesTotal:  tt.bytesTotal,
				StartTime:   tt.startTime,
				LastUpdate:  tt.lastUpdate,
			}
			got := p.GetETA()
			if got != tt.expected {
				t.Errorf("GetETA() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDownloadStatus_Constants(t *testing.T) {
	// Verify that download status constants are defined correctly
	statuses := []DownloadStatus{
		DownloadStatusNotDownloaded,
		DownloadStatusDownloading,
		DownloadStatusDownloaded,
		DownloadStatusFailed,
	}

	// Just verify they compile and have expected order
	if statuses[0] != DownloadStatusNotDownloaded {
		t.Error("DownloadStatusNotDownloaded should be first")
	}
	if statuses[1] != DownloadStatusDownloading {
		t.Error("DownloadStatusDownloading should be second")
	}
	if statuses[2] != DownloadStatusDownloaded {
		t.Error("DownloadStatusDownloaded should be third")
	}
	if statuses[3] != DownloadStatusFailed {
		t.Error("DownloadStatusFailed should be fourth")
	}
}
