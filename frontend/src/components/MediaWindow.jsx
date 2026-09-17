import React from 'react';

export default function MediaWindow({ window }) {
  const playback = window.playback;

  if (!playback || playback.is_fallback) {
    return (
      <div className="window-card">
        <div className="window-header">
          <h3 className="window-title">{window.name}</h3>
          <span className="badge badge-warning">No Media</span>
        </div>
        <div className="media-display fallback-display">
          <p>Media unavailable or playlist is empty</p>
        </div>
      </div>
    );
  }

  const { media, is_sync, remaining_seconds, elapsed_seconds } = playback;

  return (
    <div className={`window-card ${is_sync ? 'sync-active-card' : ''}`}>
      <div className="window-header">
        <h3 className="window-title">{window.name}</h3>
        {is_sync ? (
          <span className="badge badge-sync">SYNC BROADCAST</span>
        ) : (
          <span className="badge badge-normal">Playing Loop</span>
        )}
      </div>

      <div className="media-display">
        {media.media_type === 'image' && (
          <img
            src={media.media_url}
            alt={media.title}
            className="media-content"
            onError={(e) => {
              e.target.style.display = 'none';
              e.target.nextSibling.style.display = 'flex';
            }}
          />
        )}

        {media.media_type === 'video' && (
          <video
            key={media.media_url}
            src={media.media_url}
            autoPlay
            muted
            loop
            playsInline
            className="media-content"
          />
        )}

        {media.media_type === 'blank' && (
          <div className="blank-screen">
            <span className="blank-label">Quiet Interval ({media.duration_seconds}s)</span>
          </div>
        )}

        {/* Fallback box if image fails to load */}
        <div className="media-error-fallback" style={{ display: 'none' }}>
          <span>Image preview unavailable</span>
        </div>
      </div>

      <div className="window-info">
        <div className="media-info-row">
          <span className="media-title">{media.title}</span>
          <span className="media-type-tag">{media.media_type}</span>
        </div>

        <div className="progress-bar-container">
          <div
            className="progress-bar-fill"
            style={{
              width: `${Math.min(
                100,
                ((elapsed_seconds) / (media.duration_seconds || 1)) * 100
              )}%`,
            }}
          />
        </div>

        <div className="timing-info">
          <span>{elapsed_seconds}s elapsed</span>
          <span className="remaining-text">{remaining_seconds}s remaining</span>
        </div>
      </div>

      {window.playlist && window.playlist.length > 0 && (
        <div className="playlist-summary">
          <div className="playlist-header">Playlist ({window.playlist.length} items):</div>
          <ul className="playlist-list">
            {window.playlist.map((item, idx) => (
              <li
                key={item.id}
                className={`playlist-item ${
                  !is_sync && playback.current_index === idx ? 'active-item' : ''
                }`}
              >
                <span>{item.title}</span>
                <span className="item-dur">{item.duration_seconds}s</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
