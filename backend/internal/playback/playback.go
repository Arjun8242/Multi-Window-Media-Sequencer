package playback

import "media-sequencer/internal/models"

const CycleDurationSeconds = 18000 // 5 hours

// Calculate computes what media item a window should be showing right now.
// rawElapsedSeconds is the time since the window was created.
// totalPausedSeconds is the accumulated time the window spent frozen in sync broadcasts.
func Calculate(playlist []models.Media, rawElapsedSeconds int, totalPausedSeconds int) models.PlaybackResult {
    if len(playlist) == 0 {
        return models.PlaybackResult{IsFallback: true}
    }

    totalPlaylistDuration := 0
    for _, m := range playlist {
        totalPlaylistDuration += m.DurationSeconds
    }
    if totalPlaylistDuration == 0 {
        return models.PlaybackResult{IsFallback: true}
    }

    effectiveElapsed := rawElapsedSeconds - totalPausedSeconds
    if effectiveElapsed < 0 {
        effectiveElapsed = 0
    }

    // Step 1: always resets to 0 exactly at every 5-hour boundary.
    cycleElapsed := effectiveElapsed % CycleDurationSeconds

    // Step 2: where inside the playlist that corresponds to.
    offset := cycleElapsed % totalPlaylistDuration

    // Walk the playlist to find which item "offset" seconds lands on.
    runningTotal := 0
    for i, m := range playlist {
        if offset < runningTotal+m.DurationSeconds {
            elapsedInItem := offset - runningTotal
            remaining := m.DurationSeconds - elapsedInItem

            // Clamp so nothing bleeds past the 5-hour boundary.
            secondsUntilCycleEnd := CycleDurationSeconds - cycleElapsed
            if remaining > secondsUntilCycleEnd {
                remaining = secondsUntilCycleEnd
            }

            return models.PlaybackResult{
                Media:            m,
                RemainingSeconds: remaining,
                ElapsedSeconds:   elapsedInItem,
                CurrentIndex:     i,
            }
        }
        runningTotal += m.DurationSeconds
    }

    // Fail safe to item 0 if iteration finishes (shouldn't happen with correct math).
    return models.PlaybackResult{Media: playlist[0], CurrentIndex: 0}
}
