-- Schema for the URL shortener. This mirrors the in-memory `entry` struct:
-- one row per short code, holding its URL and redirect count.
CREATE TABLE IF NOT EXISTS urls (
    code  TEXT PRIMARY KEY,           -- the short code (was the map key)
    url   TEXT NOT NULL,              -- the destination URL
    count INTEGER NOT NULL DEFAULT 0  -- redirect clicks; starts at 0 like the struct
);
