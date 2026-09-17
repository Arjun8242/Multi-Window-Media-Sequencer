package playback

import (
    "testing"
    "media-sequencer/internal/models"
)

func TestBoundaryRestart(t *testing.T) {
    playlist := []models.Media{
        {DurationSeconds: 10},
        {DurationSeconds: 7},
    }

    tests := []struct {
        name             string
        elapsed          int
        paused           int
        expectedIndex    int
        expectedElapsed  int
        expectedRemain   int
    }{
        {"Start of cycle", 0, 0, 0, 0, 10},
        {"Middle of first item", 5, 0, 0, 5, 5},
        {"Transition to second item", 10, 0, 1, 0, 7},
        {"Exact 5-hour boundary restart", 18000, 0, 0, 0, 10},
        {"1 second before 5-hour boundary", 17999, 0, 1, 3, 1}, // Clamped remaining time
        {"1 second after 5-hour boundary", 18001, 0, 0, 1, 9},
        {"Second 5-hour boundary", 36000, 0, 0, 0, 10},
        {"With paused time offset", 18005, 5, 0, 0, 10}, // Effectively 18000
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := Calculate(playlist, tc.elapsed, tc.paused)
            if result.CurrentIndex != tc.expectedIndex {
                t.Errorf("Expected index %d, got %d", tc.expectedIndex, result.CurrentIndex)
            }
            if result.ElapsedSeconds != tc.expectedElapsed {
                t.Errorf("Expected elapsed %d, got %d", tc.expectedElapsed, result.ElapsedSeconds)
            }
            if result.RemainingSeconds != tc.expectedRemain {
                t.Errorf("Expected remaining %d, got %d", tc.expectedRemain, result.RemainingSeconds)
            }
        })
    }
}
