package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "media-sequencer/internal/database"
    "media-sequencer/internal/models"
    "media-sequencer/internal/playback"
)

func NewRouter(db *sql.DB) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    })

    mux.HandleFunc("GET /api/windows", func(w http.ResponseWriter, r *http.Request) {
        windows, err := database.GetWindows(db)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to load windows")
            return
        }

        sync, _ := database.GetSyncState(db)
        if sync != nil {
            elapsed := int(time.Since(sync.StartedAt).Seconds())
            if elapsed >= sync.DurationSeconds {
                // Sync naturally expired. Credit full duration to paused time.
                database.AddPausedSecondsToWindows(db, sync.DurationSeconds)
                database.ClearSyncState(db)
                sync = nil
                // Reload windows since paused_seconds changed
                windows, _ = database.GetWindows(db)
            }
        }

        // Fetch media item for sync override if active
        var syncMedia *models.Media
        if sync != nil {
            syncMedia, _ = database.GetMediaByID(db, sync.MediaID)
        }

        for i := range windows {
            playlist, _ := database.GetPlaylist(db, windows[i].ID)
            windows[i].Playlist = playlist
            
            // Raw elapsed time since the window was created
            rawElapsed := int(time.Since(windows[i].CreatedAt).Seconds())
            
            result := playback.Calculate(playlist, rawElapsed, windows[i].PausedSeconds)
            
            if sync != nil && syncMedia != nil {
                result = applySyncOverride(sync, syncMedia, result)
            }
            
            windows[i].Playback = &result
        }
        writeJSON(w, http.StatusOK, windows)
    })

    mux.HandleFunc("GET /api/windows/{id}", func(w http.ResponseWriter, r *http.Request) {
        id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
        // Simplified for this implementation - just reuses the windows list logic
        windows, err := database.GetWindows(db)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to load windows")
            return
        }
        var targetWindow *models.Window
        for i := range windows {
            if windows[i].ID == id {
                targetWindow = &windows[i]
                break
            }
        }
        if targetWindow == nil {
            writeError(w, http.StatusNotFound, "window not found")
            return
        }
        playlist, _ := database.GetPlaylist(db, targetWindow.ID)
        targetWindow.Playlist = playlist
        writeJSON(w, http.StatusOK, targetWindow)
    })

    mux.HandleFunc("POST /api/windows/{id}/media", func(w http.ResponseWriter, r *http.Request) {
        id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
        var input struct {
            Title           string `json:"title"`
            MediaType       string `json:"media_type"`
            MediaURL        string `json:"media_url"`
            DurationSeconds int    `json:"duration_seconds"`
        }
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
            writeError(w, http.StatusBadRequest, "invalid body")
            return
        }
        
        if input.Title == "" || input.DurationSeconds <= 0 {
            writeError(w, http.StatusBadRequest, "invalid inputs")
            return
        }

        created, err := database.AddMedia(db, models.Media{
            WindowID:        id,
            Title:           input.Title,
            MediaType:       input.MediaType,
            MediaURL:        input.MediaURL,
            DurationSeconds: input.DurationSeconds,
        })
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to add media")
            return
        }
        writeJSON(w, http.StatusCreated, created)
    })

    mux.HandleFunc("GET /api/media", func(w http.ResponseWriter, r *http.Request) {
        mediaList, err := database.GetAllMedia(db)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to get media list")
            return
        }
        writeJSON(w, http.StatusOK, mediaList)
    })

    mux.HandleFunc("GET /api/sync", func(w http.ResponseWriter, r *http.Request) {
        sync, _ := database.GetSyncState(db)
        if sync != nil {
            elapsed := int(time.Since(sync.StartedAt).Seconds())
            if elapsed >= sync.DurationSeconds {
                database.AddPausedSecondsToWindows(db, sync.DurationSeconds)
                database.ClearSyncState(db)
                sync = nil
            }
        }
        
        response := map[string]any{
            "is_active": sync != nil,
        }
        if sync != nil {
            remaining := sync.DurationSeconds - int(time.Since(sync.StartedAt).Seconds())
            if remaining < 0 {
                remaining = 0
            }
            response["sync_state"] = sync
            response["remaining_seconds"] = remaining
        }
        
        writeJSON(w, http.StatusOK, response)
    })

    mux.HandleFunc("POST /api/sync", func(w http.ResponseWriter, r *http.Request) {
        var input struct {
            MediaID         int64 `json:"media_id"`
            DurationSeconds int   `json:"duration_seconds"`
        }
        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
            writeError(w, http.StatusBadRequest, "invalid body")
            return
        }

        // Stop existing sync if any and credit time before starting a new one
        sync, _ := database.GetSyncState(db)
        if sync != nil {
            elapsed := int(time.Since(sync.StartedAt).Seconds())
            if elapsed < sync.DurationSeconds {
                database.AddPausedSecondsToWindows(db, elapsed)
            } else {
                database.AddPausedSecondsToWindows(db, sync.DurationSeconds)
            }
            database.ClearSyncState(db)
        }

        // If duration is not provided, use the media's natural duration
        if input.DurationSeconds <= 0 {
            media, err := database.GetMediaByID(db, input.MediaID)
            if err == nil && media != nil {
                input.DurationSeconds = media.DurationSeconds
            } else {
                input.DurationSeconds = 10 // fallback
            }
        }

        err := database.SetSyncState(db, input.MediaID, input.DurationSeconds)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "failed to start sync")
            return
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
    })

    mux.HandleFunc("POST /api/sync/clear", func(w http.ResponseWriter, r *http.Request) {
        sync, _ := database.GetSyncState(db)
        if sync != nil {
            elapsed := int(time.Since(sync.StartedAt).Seconds())
            if elapsed < sync.DurationSeconds {
                database.AddPausedSecondsToWindows(db, elapsed)
            } else {
                database.AddPausedSecondsToWindows(db, sync.DurationSeconds)
            }
            database.ClearSyncState(db)
        }
        writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
    })

    return withCORS(mux)
}

func applySyncOverride(sync *models.SyncState, syncMedia *models.Media, original models.PlaybackResult) models.PlaybackResult {
    elapsed := int(time.Since(sync.StartedAt).Seconds())
    remaining := sync.DurationSeconds - elapsed
    if remaining < 0 {
        remaining = 0
    }

    return models.PlaybackResult{
        Media:            *syncMedia,
        IsSync:           true,
        IsFallback:       false,
        RemainingSeconds: remaining,
        ElapsedSeconds:   elapsed,
        CurrentIndex:     -1, // Indicates it's not part of the normal playlist
    }
}

func withCORS(next http.Handler) http.Handler {
    allowed := os.Getenv("ALLOWED_ORIGINS")
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        if allowed == "" || allowed == "*" {
            if origin != "" {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            } else {
                w.Header().Set("Access-Control-Allow-Origin", "*")
            }
        } else {
            matched := false
            for _, o := range strings.Split(allowed, ",") {
                if strings.TrimSpace(o) == origin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    matched = true
                    break
                }
            }
            if !matched {
                // Fallback to first configured origin
                first := strings.TrimSpace(strings.Split(allowed, ",")[0])
                w.Header().Set("Access-Control-Allow-Origin", first)
            }
        }

        w.Header().Set("Vary", "Origin")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}
