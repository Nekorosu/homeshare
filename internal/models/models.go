package models

import "time"

type AdminUser struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	TOTPSecret   string     `json:"-"`
	TOTPEnabled  bool       `json:"totp_enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

type Person struct {
	ID                     int64     `json:"id"`
	Label                  string    `json:"label"`
	Notes                  string    `json:"notes,omitempty"`
	Enabled                bool      `json:"enabled"`
	StorageQuotaBytes      int64     `json:"storage_quota_bytes"`
	MonthlyUploadLimitBytes int64    `json:"monthly_upload_limit_bytes"`
	MonthlyDownloadLimitBytes int64  `json:"monthly_download_limit_bytes"`
	MaxFileSizeBytes       int64     `json:"max_file_size_bytes"`
	MaxConcurrentUploads   int       `json:"max_concurrent_uploads"`
	AllowUserKeepForever   bool      `json:"allow_user_keep_forever"`
	SessionIdleDays        int       `json:"session_idle_days"`
	SessionAbsoluteDays    int       `json:"session_absolute_days"`
	IgnoreTrafficQuota     bool      `json:"ignore_traffic_quota"`
	CreatedAt              time.Time `json:"created_at"`
	LastActivityAt         *time.Time `json:"last_activity_at,omitempty"`
}

type InviteCode struct {
	ID               int64      `json:"id"`
	PersonID         int64      `json:"person_id"`
	CodeHash         string     `json:"-"`
	Enabled          bool       `json:"enabled"`
	MaxActivations   int        `json:"max_activations"`
	ActivationsUsed  int        `json:"activations_used"`
	ExpiresAt        time.Time  `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
	CreatedByAdminID *int64     `json:"created_by_admin_id,omitempty"`
	Code             string     `json:"code,omitempty"` // Показывается только при создании
}

type DeviceSession struct {
	ID                 int64      `json:"id"`
	PersonID           int64      `json:"person_id"`
	Name               string     `json:"name,omitempty"`
	SessionTokenHash   string     `json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	LastIPHash         string     `json:"last_ip_hash,omitempty"`
	LastUserAgentHash  string     `json:"-"`
	IdleExpiresAt      time.Time  `json:"idle_expires_at"`
	AbsoluteExpiresAt  time.Time  `json:"absolute_expires_at"`
	Revoked            bool       `json:"revoked"`
}

type UploadStatus string

const (
	UploadStatusReserved   UploadStatus = "reserved"
	UploadStatusUploading  UploadStatus = "uploading"
	UploadStatusCompleted  UploadStatus = "completed"
	UploadStatusCanceled   UploadStatus = "canceled"
	UploadStatusExpired    UploadStatus = "expired"
	UploadStatusFailed     UploadStatus = "failed"
)

type Upload struct {
	ID                   int64        `json:"id"`
	PersonID             int64        `json:"person_id"`
	SessionID            int64        `json:"session_id"`
	UploadSecretHash     string       `json:"-"`
	OriginalName         string       `json:"original_name"`
	DeclaredSize         int64        `json:"declared_size"`
	ReceivedBytes        int64        `json:"received_bytes"`
	Status               UploadStatus `json:"status"`
	ExpiryDays           int          `json:"expiry_days"`
	ReservationExpiresAt time.Time    `json:"reservation_expires_at"`
	CreatedAt            time.Time    `json:"created_at"`
	CompletedAt          *time.Time   `json:"completed_at,omitempty"`
	ClientIPHash         string       `json:"client_ip_hash,omitempty"`
	UploadSecret         string       `json:"upload_secret,omitempty"` // Показывается только при создании
}

type FileStatus string

const (
	FileStatusReady       FileStatus = "ready"
	FileStatusQuarantined FileStatus = "quarantined"
)

type File struct {
	ID           string     `json:"id"`
	PersonID     int64      `json:"person_id"`
	OriginalName string     `json:"original_name"`
	StoredPath   string     `json:"-"`
	Size         int64      `json:"size"`
	ContentType  string     `json:"content_type,omitempty"`
	Status       FileStatus `json:"status"`
	Flagged      bool       `json:"flagged"`
	FlagReason   string     `json:"flag_reason,omitempty"`
	Protected    bool       `json:"protected"`
	KeepForever  bool       `json:"keep_forever"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ClientIPHash string     `json:"client_ip_hash,omitempty"`
	UploaderLabel string    `json:"uploader_label,omitempty"` // Snapshot label
}

type TrafficCounter struct {
	PersonID              int64     `json:"person_id"`
	Month                 string    `json:"month"`
	UploadCompletedBytes  int64     `json:"upload_completed_bytes"`
	UploadAbortedBytes    int64     `json:"upload_aborted_bytes"`
	DownloadCompletedBytes int64    `json:"download_completed_bytes"`
	DownloadAbortedBytes  int64     `json:"download_aborted_bytes"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID         int64      `json:"id"`
	Time       time.Time  `json:"time"`
	ActorType  string     `json:"actor_type"`
	ActorID    *int64     `json:"actor_id,omitempty"`
	Event      string     `json:"event"`
	EntityType string     `json:"entity_type,omitempty"`
	EntityID   string     `json:"entity_id,omitempty"`
	IPHash     string     `json:"ip_hash,omitempty"`
	Details    string     `json:"details,omitempty"`
}

type RateLimitLock struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DashboardStats struct {
	TotalDiskBytes       int64   `json:"total_disk_bytes"`
	FreeDiskBytes        int64   `json:"free_disk_bytes"`
	UsedStorageBytes     int64   `json:"used_storage_bytes"`
	ExternalUploadBps    float64 `json:"external_upload_bps"`
	ExternalDownloadBps  float64 `json:"external_download_bps"`
	ActiveExternalUploads int64  `json:"active_external_uploads"`
	ActiveSessions       int64   `json:"active_sessions"`
	QuarantinedFiles     int64   `json:"quarantined_files"`
}
