-- SQLite schema для homeshare

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;

-- AdminUser
CREATE TABLE IF NOT EXISTS admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    totp_enabled BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME
);

-- Person
CREATE TABLE IF NOT EXISTS persons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT NOT NULL,
    notes TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    storage_quota_bytes INTEGER DEFAULT 107374182400,
    monthly_upload_limit_bytes INTEGER DEFAULT 214748364800,
    monthly_download_limit_bytes INTEGER DEFAULT 322122547200,
    max_file_size_bytes INTEGER DEFAULT 53687091200,
    max_concurrent_uploads INTEGER DEFAULT 1,
    allow_user_keep_forever BOOLEAN DEFAULT FALSE,
    session_idle_days INTEGER DEFAULT 30,
    session_absolute_days INTEGER DEFAULT 90,
    ignore_traffic_quota BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity_at DATETIME
);

-- InviteCode
CREATE TABLE IF NOT EXISTS invite_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    code_hash TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    max_activations INTEGER DEFAULT 1,
    activations_used INTEGER DEFAULT 0,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by_admin_id INTEGER,
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by_admin_id) REFERENCES admin_users(id)
);

-- DeviceSession
CREATE TABLE IF NOT EXISTS device_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    name TEXT,
    session_token_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    last_ip_hash TEXT,
    last_user_agent_hash TEXT,
    idle_expires_at DATETIME NOT NULL,
    absolute_expires_at DATETIME NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE
);

-- Upload
CREATE TABLE IF NOT EXISTS uploads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    person_id INTEGER NOT NULL,
    session_id INTEGER NOT NULL,
    upload_secret_hash TEXT NOT NULL,
    original_name TEXT NOT NULL,
    declared_size INTEGER NOT NULL,
    received_bytes INTEGER DEFAULT 0,
    status TEXT DEFAULT 'reserved' CHECK(status IN ('reserved', 'uploading', 'completed', 'canceled', 'expired', 'failed')),
    expiry_days INTEGER DEFAULT 14,
    reservation_expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    client_ip_hash TEXT,
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES device_sessions(id) ON DELETE CASCADE
);

-- File
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    person_id INTEGER NOT NULL,
    original_name TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    size INTEGER NOT NULL,
    content_type TEXT,
    status TEXT DEFAULT 'ready' CHECK(status IN ('ready', 'quarantined')),
    flagged BOOLEAN DEFAULT FALSE,
    flag_reason TEXT,
    protected BOOLEAN DEFAULT FALSE,
    keep_forever BOOLEAN DEFAULT FALSE,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    client_ip_hash TEXT,
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE SET NULL
);

-- TrafficCounter
CREATE TABLE IF NOT EXISTS traffic_counters (
    person_id INTEGER NOT NULL,
    month TEXT NOT NULL,
    upload_completed_bytes INTEGER DEFAULT 0,
    upload_aborted_bytes INTEGER DEFAULT 0,
    download_completed_bytes INTEGER DEFAULT 0,
    download_aborted_bytes INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (person_id, month),
    FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE CASCADE
);

-- AuditLog
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    time DATETIME DEFAULT CURRENT_TIMESTAMP,
    actor_type TEXT CHECK(actor_type IN ('admin', 'person', 'system')),
    actor_id INTEGER,
    event TEXT NOT NULL,
    entity_type TEXT,
    entity_id TEXT,
    ip_hash TEXT,
    details TEXT
);

-- RateLimitLock
CREATE TABLE IF NOT EXISTS rate_limit_locks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    reason TEXT,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Settings
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_persons_enabled ON persons(enabled);
CREATE INDEX IF NOT EXISTS idx_invite_codes_code_hash ON invite_codes(code_hash);
CREATE INDEX IF NOT EXISTS idx_invite_codes_expires ON invite_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_device_sessions_token ON device_sessions(session_token_hash);
CREATE INDEX IF NOT EXISTS idx_device_sessions_person ON device_sessions(person_id);
CREATE INDEX IF NOT EXISTS idx_uploads_status ON uploads(status);
CREATE INDEX IF NOT EXISTS idx_uploads_reservation_expires ON uploads(reservation_expires_at);
CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
CREATE INDEX IF NOT EXISTS idx_files_expires ON files(expires_at);
CREATE INDEX IF NOT EXISTS idx_files_person ON files(person_id);
CREATE INDEX IF NOT EXISTS idx_traffic_month ON traffic_counters(month);
CREATE INDEX IF NOT EXISTS idx_audit_time ON audit_log(time);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_log(actor_type, actor_id);
CREATE INDEX IF NOT EXISTS idx_rate_limit_expires ON rate_limit_locks(expires_at);

-- Начальные настройки
INSERT OR IGNORE INTO settings (key, value) VALUES 
    ('default_storage_quota_bytes', '107374182400'),
    ('default_monthly_upload_bytes', '214748364800'),
    ('default_monthly_download_bytes', '322122547200'),
    ('default_max_file_size_bytes', '53687091200'),
    ('default_max_concurrent_uploads', '1'),
    ('external_upload_limit_mbps', '250'),
    ('external_download_limit_mbps', '250'),
    ('burst_mb', '16'),
    ('zip_max_files', '100'),
    ('zip_max_total_bytes', '53687091200'),
    ('quarantine_suspicious', 'true'),
    ('default_expiry_days', '14');
