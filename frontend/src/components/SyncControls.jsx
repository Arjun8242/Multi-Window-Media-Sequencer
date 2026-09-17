import React, { useState } from 'react';

export default function SyncControls({ syncData, mediaList, onStartSync, onClearSync }) {
  const [selectedMediaId, setSelectedMediaId] = useState('');
  const [customDuration, setCustomDuration] = useState('');
  const [loading, setLoading] = useState(false);

  const isActive = syncData?.is_active;
  const remainingSeconds = syncData?.remaining_seconds ?? 0;

  const handleStart = async (e) => {
    e.preventDefault();
    if (!selectedMediaId) return;
    setLoading(true);
    try {
      await onStartSync(selectedMediaId, customDuration || undefined);
    } finally {
      setLoading(false);
    }
  };

  const handleClear = async () => {
    setLoading(true);
    try {
      await onClearSync();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card control-card">
      <div className="card-header">
        <h2>Sync Playback Controls</h2>
        {isActive ? (
          <span className="badge badge-active-pulse">
            Active Broadcast: {remainingSeconds}s left
          </span>
        ) : (
          <span className="badge badge-idle">Idle (Independent Loops)</span>
        )}
      </div>

      <p className="description-text">
        Broadcast a single media asset simultaneously across all windows. When broadcast ends or is cleared, each window resumes its previous position without jumping ahead.
      </p>

      {isActive && (
        <div className="active-sync-banner">
          <div className="active-sync-info">
            <strong>Active Sync:</strong> Media #{syncData.sync_state?.media_id} (Duration: {syncData.sync_state?.duration_seconds}s)
          </div>
          <button
            type="button"
            className="btn btn-danger"
            onClick={handleClear}
            disabled={loading}
          >
            Clear / Resume Normal
          </button>
        </div>
      )}

      <form onSubmit={handleStart} className="sync-form">
        <div className="form-group">
          <label htmlFor="media-select">Select Media to Sync</label>
          <select
            id="media-select"
            className="form-input"
            value={selectedMediaId}
            onChange={(e) => setSelectedMediaId(e.target.value)}
            required
          >
            <option value="">-- Choose media from any window --</option>
            {mediaList.map((m) => (
              <option key={m.id} value={m.id}>
                #{m.id} - {m.title} ({m.media_type}, {m.duration_seconds}s) [Window {m.window_id}]
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="custom-duration">Duration in Seconds (Optional)</label>
          <input
            id="custom-duration"
            type="number"
            min="1"
            className="form-input"
            placeholder="Leave empty to use natural duration"
            value={customDuration}
            onChange={(e) => setCustomDuration(e.target.value)}
          />
        </div>

        <div className="button-row">
          <button
            type="submit"
            className="btn btn-primary"
            disabled={loading || !selectedMediaId}
          >
            {loading ? 'Broadcasting...' : 'Broadcast to All Windows'}
          </button>
          {isActive && (
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClear}
              disabled={loading}
            >
              Clear Sync
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
