import React, { useEffect, useState, useCallback } from 'react';
import { getWindows, getSync, getAllMedia, startSync, clearSync, addMedia } from './services/api';
import MediaWindow from './components/MediaWindow';
import SyncControls from './components/SyncControls';
import AddMediaForm from './components/AddMediaForm';

export default function App() {
  const [windows, setWindows] = useState([]);
  const [syncData, setSyncData] = useState(null);
  const [mediaList, setMediaList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdated, setLastUpdated] = useState(null);

  const fetchData = useCallback(async () => {
    try {
      const [windowsData, syncRes, mediaRes] = await Promise.all([
        getWindows(),
        getSync(),
        getAllMedia(),
      ]);
      setWindows(windowsData || []);
      setSyncData(syncRes || null);
      setMediaList(mediaRes || []);
      setLastUpdated(new Date().toLocaleTimeString());
      setError(null);
    } catch (err) {
      console.error('Fetch error:', err);
      setError('Unable to connect to backend service (http://localhost:8080). Make sure the Go backend is running.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    // 2-second polling interval as specified in task instructions
    const interval = setInterval(fetchData, 2000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const handleStartSync = async (mediaId, duration) => {
    await startSync(mediaId, duration);
    await fetchData();
  };

  const handleClearSync = async () => {
    await clearSync();
    await fetchData();
  };

  const handleAddMedia = async (windowId, payload) => {
    await addMedia(windowId, payload);
    await fetchData();
  };

  return (
    <div className="app-container">
      <header className="app-header">
        <div className="header-left">
          <div className="logo-icon">📺</div>
          <div>
            <h1>Multi-Window Media Sequencer</h1>
            <p className="subtitle">Synchronized Multi-Display Playback Engine with 5-Hour Modulo Boundary</p>
          </div>
        </div>
        <div className="header-right">
          <div className="status-indicator">
            <span className={`status-dot ${error ? 'dot-offline' : 'dot-online'}`}></span>
            <span>{error ? 'Backend Offline' : 'Engine Connected'}</span>
          </div>
          {lastUpdated && <span className="timestamp">Synced: {lastUpdated}</span>}
          <button className="btn-refresh" onClick={fetchData} title="Refresh immediately">
            🔄
          </button>
        </div>
      </header>

      {error && (
        <div className="banner-alert">
          <strong>Connection Warning: </strong> {error}
        </div>
      )}

      <main className="main-content">
        <section className="windows-section">
          <div className="section-title-row">
            <h2>Active Display Windows ({windows.length})</h2>
            <span className="info-tag">Auto-refreshing every 2s</span>
          </div>

          {loading && windows.length === 0 ? (
            <div className="loading-state">
              <div className="spinner"></div>
              <p>Connecting to media engine...</p>
            </div>
          ) : (
            <div className="window-grid">
              {windows.map((win) => (
                <MediaWindow key={win.id} window={win} />
              ))}
            </div>
          )}
        </section>

        <section className="controls-section">
          <div className="controls-grid">
            <SyncControls
              syncData={syncData}
              mediaList={mediaList}
              onStartSync={handleStartSync}
              onClearSync={handleClearSync}
            />
            <AddMediaForm
              windows={windows}
              onMediaAdded={handleAddMedia}
            />
          </div>
        </section>
      </main>

      <footer className="app-footer">
        <div className="footer-content">
          <span>Multi-Window Media Sequencer — 5-Hour Cycle Modulo (18,000s) Engine</span>
          <span>PostgreSQL + Go Backend | React Frontend</span>
        </div>
      </footer>
    </div>
  );
}
