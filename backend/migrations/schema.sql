CREATE TABLE IF NOT EXISTS windows (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paused_seconds INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS media (
    id BIGSERIAL PRIMARY KEY,
    window_id BIGINT NOT NULL REFERENCES windows(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    media_type VARCHAR(20) NOT NULL CHECK (media_type IN ('image', 'video', 'blank')),
    media_url TEXT NOT NULL DEFAULT '',
    duration_seconds INT NOT NULL CHECK (duration_seconds > 0),
    display_order INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_media_url CHECK (media_type = 'blank' OR length(trim(media_url)) > 0)
);
CREATE INDEX IF NOT EXISTS idx_media_window_order ON media(window_id, display_order);

CREATE TABLE IF NOT EXISTS sync_state (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    media_id BIGINT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    duration_seconds INT NOT NULL CHECK (duration_seconds > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- seed data
INSERT INTO windows (id, name) VALUES
    (1, 'Window 1 - Main Display'),
    (2, 'Window 2 - Promo Screen'),
    (3, 'Window 3 - Side Banner')
ON CONFLICT (id) DO NOTHING;

-- Reset sequence if needed after explicit id insertion
SELECT setval('windows_id_seq', (SELECT COALESCE(MAX(id), 1) FROM windows));

INSERT INTO media (id, window_id, title, media_type, media_url, duration_seconds, display_order) VALUES
    (1, 1, 'Welcome Image', 'image', 'https://picsum.photos/id/1015/1280/720', 10, 1),
    (2, 1, 'Sample Clip',   'video', 'https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4', 12, 2),
    (3, 2, 'Promo Banner',  'image', 'https://picsum.photos/id/1025/1280/720', 8, 1),
    (4, 2, 'Quiet Interval','blank', '', 5, 2),
    (5, 3, 'Notice',        'image', 'https://picsum.photos/id/1035/1280/720', 10, 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval('media_id_seq', (SELECT COALESCE(MAX(id), 1) FROM media));
