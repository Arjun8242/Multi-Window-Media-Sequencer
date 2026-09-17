# Multi-Window Media Sequencer with Sync Playback

A full-stack, synchronized multi-display digital signage media player built with **Go**, **PostgreSQL**, and **React**.

The application orchestrates independent repeating playlists across multiple windows, strictly enforces an exact 5-hour (18,000s) boundary reset, and provides global sync broadcast with freeze-and-resume behavior.

---

## 1. System Architecture

```
┌─────────────────────────┐          HTTP (JSON)          ┌──────────────────────────┐         SQL         ┌─────────────────────────┐
│     React Frontend      │ ───────────────────────────▶  │      Go REST Backend     │ ──────────────────▶ │   PostgreSQL Database   │
│                         │ ◀───────────────────────────  │                          │ ◀────────────────── │                         │
│  - Multi-window display │       polls every 2s          │  - Stateless playback    │                     │  - windows table        │
│  - Media renderers      │                               │    calculation           │                     │  - media table          │
│  - Sync controls        │                               │  - Sync freeze/resume    │                     │  - sync_state singleton │
│  - Playlist manager     │                               │  - 5-hour modulo math    │                     │                         │
└─────────────────────────┘                               └──────────────────────────┘                     └─────────────────────────┘
```

### Core Playback Math: The 5-Hour Boundary
Each window's playback is calculated statelessly on demand using two nested modulos:

1. **5-Hour Cycle Clock**:
   ```
   cycleElapsed = (rawElapsedSeconds - totalPausedSeconds) % 18000
   ```
   Because `18000` seconds equals 5 hours, `cycleElapsed` resets to `0` at exactly every 5-hour mark (`0`, `18000`, `36000`, ...), regardless of individual media durations.

2. **Playlist Offset**:
   ```
   offset = cycleElapsed % totalPlaylistDuration
   ```
   Walking forward through playlist item durations identifies which media is active and computes remaining seconds clamped to the 5-hour boundary:
   ```go
   secondsUntilCycleEnd := CycleDurationSeconds - cycleElapsed
   if remaining > secondsUntilCycleEnd {
       remaining = secondsUntilCycleEnd
   }
   ```

---

## 2. Sync Playback & Freeze/Resume Behavior

Sync broadcasts allow an operator to temporarily display a chosen media item on **all** windows simultaneously.

1. **Broadcast Trigger**:
   - Operator selects a media item and submits `POST /api/sync`.
   - Backend records active state in the `sync_state` table with `started_at` timestamp.
2. **During Sync**:
   - All windows temporarily override their display to show the synced media item and countdown.
3. **Freeze & Resume (No Skipping)**:
   - While sync is running, the time windows spend frozen is tracked.
   - When sync ends naturally (elapsed time reaches duration) or manually via `POST /api/sync/clear`, the elapsed freeze duration is credited to each window's `paused_seconds` in the database.
   - In subsequent calculations, `(rawElapsedSeconds - pausedSeconds)` ensures that when windows return to their normal playlist, they pick up right where they left off instead of skipping ahead.

---

## 3. Database Schema

- **`windows`**:
  - `id`: BIGSERIAL PRIMARY KEY
  - `name`: VARCHAR(100) NOT NULL
  - `created_at`: TIMESTAMPTZ NOT NULL DEFAULT NOW()
  - `paused_seconds`: INT NOT NULL DEFAULT 0 (tracks frozen time during syncs)
- **`media`**:
  - `id`: BIGSERIAL PRIMARY KEY
  - `window_id`: BIGINT NOT NULL REFERENCES windows(id) ON DELETE CASCADE
  - `title`: VARCHAR(150) NOT NULL
  - `media_type`: VARCHAR(20) CHECK (media_type IN ('image', 'video', 'blank'))
  - `media_url`: TEXT NOT NULL DEFAULT ''
  - `duration_seconds`: INT NOT NULL CHECK (duration_seconds > 0)
  - `display_order`: INT NOT NULL DEFAULT 1
  - `created_at`: TIMESTAMPTZ NOT NULL DEFAULT NOW()
- **`sync_state`**:
  - `id`: INT PRIMARY KEY DEFAULT 1 CHECK (id = 1) (Enforces singleton pattern)
  - `media_id`: BIGINT NOT NULL REFERENCES media(id) ON DELETE CASCADE
  - `started_at`: TIMESTAMPTZ NOT NULL
  - `duration_seconds`: INT NOT NULL CHECK (duration_seconds > 0)
  - `updated_at`: TIMESTAMPTZ NOT NULL DEFAULT NOW()

---

## 4. API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health and liveness check (`{"status":"ok"}`) |
| `GET` | `/api/windows` | Returns all windows, playlists, and current computed playback state |
| `GET` | `/api/windows/{id}` | Returns a single window with its playlist |
| `POST` | `/api/windows/{id}/media` | Appends a new media item to a window's playlist |
| `GET` | `/api/media` | Returns all media items across all windows (for sync selection) |
| `GET` | `/api/sync` | Returns current sync status and remaining countdown |
| `POST` | `/api/sync` | Starts a sync broadcast (`{"media_id": 1, "duration_seconds": 15}`) |
| `POST` | `/api/sync/clear` | Cancels the active sync broadcast early |

---

## 5. Local Setup & Running Instructions

### Prerequisites
- **Go** (1.20+)
- **PostgreSQL** (running locally, default port `5432` or custom like `5433`)
- **Node.js** (v18+) & **npm**

### Step 1: Database Setup
Create the database and execute the migration schema:
```bash
# In PostgreSQL terminal (psql)
CREATE DATABASE mediadb;
\c mediadb
\i backend/migrations/schema.sql
```
*(Note: The Go backend also automatically checks and runs `schema.sql` on startup if tables do not exist).*

### Step 2: Backend Configuration & Launch
Configure `backend/.env`:
```env
DATABASE_URL=postgres://postgres:your_password@localhost:5433/mediadb?sslmode=disable
PORT=8080
ALLOWED_ORIGINS=http://localhost:5173
```
Run the Go backend:
```bash
cd backend
go run ./cmd/server
```
The server will start listening on port `8080`.

### Step 3: Frontend Launch
```bash
cd frontend
npm install
npm run dev
```
Open your browser at `http://localhost:5173`.

---

## 6. Running Tests

To run the playback engine test suite verifying the 18,000s boundary calculations and offset handling:
```bash
cd backend
go test ./internal/playback/...
```

---

## 7. Deployment Instructions

1. **Database**: Provision a managed PostgreSQL instance (e.g., Supabase, Neon, Railway, or Render). Run `schema.sql` to initialize tables and seed data.
2. **Backend**:
   - Deploy `backend/` as a Web Service on Render or Railway.
   - Configure environment variables:
     - `DATABASE_URL`: Connection string of your managed database.
     - `PORT`: Service port (e.g., `8080`).
     - `ALLOWED_ORIGINS`: Production frontend URL (e.g., `https://your-frontend.vercel.app`).
3. **Frontend**:
   - Deploy `frontend/` to Vercel or Netlify.
   - Set build command: `npm run build`
   - Set publish directory: `dist`
   - Configure environment variable:
     - `VITE_API_BASE_URL`: Production backend URL (e.g., `https://your-backend.onrender.com`).
