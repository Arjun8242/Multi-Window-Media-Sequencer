const BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// Fetch all windows with current playback states
export async function getWindows() {
  const res = await fetch(`${BASE}/api/windows`);
  if (!res.ok) {
    throw new Error(`Failed to load windows: ${res.statusText}`);
  }
  return res.json();
}

// Add a media item to a specific window
export async function addMedia(windowId, payload) {
  const res = await fetch(`${BASE}/api/windows/${windowId}/media`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.error || `Failed to add media: ${res.statusText}`);
  }
  return res.json();
}

// Fetch all media items across all windows
export async function getAllMedia() {
  const res = await fetch(`${BASE}/api/media`);
  if (!res.ok) {
    throw new Error(`Failed to load media: ${res.statusText}`);
  }
  return res.json();
}

// Fetch current sync broadcast status
export async function getSync() {
  const res = await fetch(`${BASE}/api/sync`);
  if (!res.ok) {
    throw new Error(`Failed to check sync status: ${res.statusText}`);
  }
  return res.json();
}

// Broadcast a media item to all windows
export async function startSync(mediaId, durationSeconds) {
  const body = { media_id: Number(mediaId) };
  if (durationSeconds) {
    body.duration_seconds = Number(durationSeconds);
  }
  const res = await fetch(`${BASE}/api/sync`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(`Failed to trigger sync: ${res.statusText}`);
  }
  return res.json();
}

// Cancel active sync broadcast
export async function clearSync() {
  const res = await fetch(`${BASE}/api/sync/clear`, {
    method: 'POST',
  });
  if (!res.ok) {
    throw new Error(`Failed to clear sync: ${res.statusText}`);
  }
  return res.json();
}
