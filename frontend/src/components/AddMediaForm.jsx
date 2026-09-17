import React, { useState } from 'react';

export default function AddMediaForm({ windows, onMediaAdded }) {
  const [windowId, setWindowId] = useState('');
  const [title, setTitle] = useState('');
  const [mediaType, setMediaType] = useState('image');
  const [mediaUrl, setMediaUrl] = useState('');
  const [durationSeconds, setDurationSeconds] = useState(10);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage('');
    setError('');

    const targetWindowId = windowId || (windows[0]?.id ? String(windows[0].id) : '');
    if (!targetWindowId) {
      setError('Please select a target window');
      return;
    }

    if (mediaType !== 'blank' && !mediaUrl.trim()) {
      setError('Media URL is required for images and videos');
      return;
    }

    setLoading(true);
    try {
      await onMediaAdded(targetWindowId, {
        title: title.trim(),
        media_type: mediaType,
        media_url: mediaType === 'blank' ? '' : mediaUrl.trim(),
        duration_seconds: Number(durationSeconds),
      });

      setMessage(`Added "${title}" successfully!`);
      setTitle('');
      setMediaUrl('');
      setDurationSeconds(10);
    } catch (err) {
      setError(err.message || 'Failed to add media item');
    } finally {
      setLoading(false);
    }
  };

  const fillSample = (type) => {
    if (type === 'image') {
      setTitle('Sample Nature Photo');
      setMediaType('image');
      setMediaUrl('https://picsum.photos/id/1043/1280/720');
      setDurationSeconds(10);
    } else if (type === 'video') {
      setTitle('Big Buck Bunny Clip');
      setMediaType('video');
      setMediaUrl('https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4');
      setDurationSeconds(15);
    } else if (type === 'blank') {
      setTitle('Intermission Pause');
      setMediaType('blank');
      setMediaUrl('');
      setDurationSeconds(5);
    }
  };

  return (
    <div className="card control-card">
      <div className="card-header">
        <h2>Add Media to Window Playlist</h2>
      </div>

      <p className="description-text">
        Append new items to a window's playlist dynamically. The sequence will seamlessly adapt without restarting unaffected loops.
      </p>

      {message && <div className="alert alert-success">{message}</div>}
      {error && <div className="alert alert-danger">{error}</div>}

      <div className="sample-buttons">
        <span>Quick Samples: </span>
        <button type="button" className="btn-chip" onClick={() => fillSample('image')}>
          + Nature Image
        </button>
        <button type="button" className="btn-chip" onClick={() => fillSample('video')}>
          + Video Clip
        </button>
        <button type="button" className="btn-chip" onClick={() => fillSample('blank')}>
          + 5s Blank Pause
        </button>
      </div>

      <form onSubmit={handleSubmit} className="add-media-form">
        <div className="form-row">
          <div className="form-group flex-1">
            <label htmlFor="target-window">Target Window</label>
            <select
              id="target-window"
              className="form-input"
              value={windowId}
              onChange={(e) => setWindowId(e.target.value)}
            >
              {windows.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name} (ID: {w.id})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group flex-1">
            <label htmlFor="media-type">Media Type</label>
            <select
              id="media-type"
              className="form-input"
              value={mediaType}
              onChange={(e) => setMediaType(e.target.value)}
            >
              <option value="image">Image</option>
              <option value="video">Video</option>
              <option value="blank">Blank (Quiet interval)</option>
            </select>
          </div>
        </div>

        <div className="form-row">
          <div className="form-group flex-2">
            <label htmlFor="media-title">Title</label>
            <input
              id="media-title"
              type="text"
              className="form-input"
              placeholder="e.g. Summer Promo"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          <div className="form-group flex-1">
            <label htmlFor="media-duration">Duration (Seconds)</label>
            <input
              id="media-duration"
              type="number"
              min="1"
              className="form-input"
              value={durationSeconds}
              onChange={(e) => setDurationSeconds(e.target.value)}
              required
            />
          </div>
        </div>

        {mediaType !== 'blank' && (
          <div className="form-group">
            <label htmlFor="media-url">Media URL</label>
            <input
              id="media-url"
              type="url"
              className="form-input"
              placeholder="https://..."
              value={mediaUrl}
              onChange={(e) => setMediaUrl(e.target.value)}
              required
            />
          </div>
        )}

        <button type="submit" className="btn btn-primary" disabled={loading}>
          {loading ? 'Adding...' : 'Add to Playlist'}
        </button>
      </form>
    </div>
  );
}
