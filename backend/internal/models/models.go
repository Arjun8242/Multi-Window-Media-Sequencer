package models

import "time"

type Window struct {
    ID            int64           `json:"id"`
    Name          string          `json:"name"`
    CreatedAt     time.Time       `json:"created_at"`
    PausedSeconds int             `json:"paused_seconds"`
    Playlist      []Media         `json:"playlist,omitempty"`
    Playback      *PlaybackResult `json:"playback,omitempty"`
}

type Media struct {
    ID              int64     `json:"id"`
    WindowID        int64     `json:"window_id"`
    Title           string    `json:"title"`
    MediaType       string    `json:"media_type"` // image | video | blank
    MediaURL        string    `json:"media_url"`
    DurationSeconds int       `json:"duration_seconds"`
    DisplayOrder    int       `json:"display_order"`
    CreatedAt       time.Time `json:"created_at"`
}

type SyncState struct {
    ID              int       `json:"id"`
    MediaID         int64     `json:"media_id"`
    StartedAt       time.Time `json:"started_at"`
    DurationSeconds int       `json:"duration_seconds"`
}

type PlaybackResult struct {
    Media            Media `json:"media"`
    IsSync           bool  `json:"is_sync"`
    IsFallback       bool  `json:"is_fallback"`
    RemainingSeconds int   `json:"remaining_seconds"`
    ElapsedSeconds   int   `json:"elapsed_seconds"`
    CurrentIndex     int   `json:"current_index"`
}
