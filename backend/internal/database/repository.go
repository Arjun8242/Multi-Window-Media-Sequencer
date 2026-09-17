package database

import (
    "database/sql"
    "media-sequencer/internal/models"
)

func GetWindows(db *sql.DB) ([]models.Window, error) {
    rows, err := db.Query(`SELECT id, name, created_at, paused_seconds FROM windows ORDER BY id`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var windows []models.Window
    for rows.Next() {
        var w models.Window
        if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.PausedSeconds); err != nil {
            return nil, err
        }
        windows = append(windows, w)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return windows, nil
}

func GetPlaylist(db *sql.DB, windowID int64) ([]models.Media, error) {
    rows, err := db.Query(`
        SELECT id, window_id, title, media_type, media_url, duration_seconds, display_order, created_at
        FROM media WHERE window_id = $1 ORDER BY display_order`, windowID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []models.Media
    for rows.Next() {
        var m models.Media
        if err := rows.Scan(&m.ID, &m.WindowID, &m.Title, &m.MediaType, &m.MediaURL,
            &m.DurationSeconds, &m.DisplayOrder, &m.CreatedAt); err != nil {
            return nil, err
        }
        items = append(items, m)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}

func GetAllMedia(db *sql.DB) ([]models.Media, error) {
    rows, err := db.Query(`
        SELECT id, window_id, title, media_type, media_url, duration_seconds, display_order, created_at
        FROM media ORDER BY window_id, display_order`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []models.Media
    for rows.Next() {
        var m models.Media
        if err := rows.Scan(&m.ID, &m.WindowID, &m.Title, &m.MediaType, &m.MediaURL,
            &m.DurationSeconds, &m.DisplayOrder, &m.CreatedAt); err != nil {
            return nil, err
        }
        items = append(items, m)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}

func AddMedia(db *sql.DB, m models.Media) (models.Media, error) {
    err := db.QueryRow(`
        INSERT INTO media (window_id, title, media_type, media_url, duration_seconds, display_order)
        VALUES ($1, $2, $3, $4, $5,
            COALESCE((SELECT MAX(display_order) FROM media WHERE window_id = $1), 0) + 1)
        RETURNING id, created_at, display_order`,
        m.WindowID, m.Title, m.MediaType, m.MediaURL, m.DurationSeconds,
    ).Scan(&m.ID, &m.CreatedAt, &m.DisplayOrder)
    return m, err
}

func GetSyncState(db *sql.DB) (*models.SyncState, error) {
    var s models.SyncState
    err := db.QueryRow(`SELECT id, media_id, started_at, duration_seconds FROM sync_state WHERE id = 1`).
        Scan(&s.ID, &s.MediaID, &s.StartedAt, &s.DurationSeconds)
    if err == sql.ErrNoRows {
        return nil, nil // no active sync
    }
    return &s, err
}

func SetSyncState(db *sql.DB, mediaID int64, durationSeconds int) error {
    _, err := db.Exec(`
        INSERT INTO sync_state (id, media_id, started_at, duration_seconds)
        VALUES (1, $1, NOW(), $2)
        ON CONFLICT (id) DO UPDATE SET media_id = $1, started_at = NOW(), duration_seconds = $2`,
        mediaID, durationSeconds)
    return err
}

func ClearSyncState(db *sql.DB) error {
    _, err := db.Exec(`DELETE FROM sync_state WHERE id = 1`)
    return err
}

func AddPausedSecondsToWindows(db *sql.DB, pausedSeconds int) error {
    _, err := db.Exec(`UPDATE windows SET paused_seconds = paused_seconds + $1`, pausedSeconds)
    return err
}

func GetMediaByID(db *sql.DB, mediaID int64) (*models.Media, error) {
    var m models.Media
    err := db.QueryRow(`
        SELECT id, window_id, title, media_type, media_url, duration_seconds, display_order, created_at
        FROM media WHERE id = $1`, mediaID).
        Scan(&m.ID, &m.WindowID, &m.Title, &m.MediaType, &m.MediaURL, &m.DurationSeconds, &m.DisplayOrder, &m.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &m, nil
}
