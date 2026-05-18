CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO settings(key, value) VALUES ('late_threshold', '09:30')
ON CONFLICT (key) DO NOTHING;

