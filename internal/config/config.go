package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen             string         `yaml:"listen"`
	BaseURL            string         `yaml:"base_url"`
	DataDir            string         `yaml:"data_dir"`
	TmpDir             string         `yaml:"tmp_dir"`
	DBPath             string         `yaml:"db_path"`
	BackupDir          string         `yaml:"backup_dir"`
	LocalCIDRs         []string       `yaml:"local_cidrs"`
	Log                LogConfig      `yaml:"log"`
	SpeedLimits        SpeedConfig    `yaml:"speed_limits"`
	Quotas             QuotaConfig    `yaml:"quotas"`
	Traffic            TrafficConfig  `yaml:"traffic"`
	RateLimits         RateLimitConfig `yaml:"rate_limits"`
	SuspiciousExtensions []string     `yaml:"suspicious_extensions"`
	Zip                ZipConfig      `yaml:"zip"`
	Sessions           SessionConfig  `yaml:"sessions"`
	Invite             InviteConfig   `yaml:"invite"`
	Disk               DiskConfig     `yaml:"disk"`
	Upload             UploadConfig   `yaml:"upload"`
	Security           SecurityConfig `yaml:"security"`
	Cleanup            CleanupConfig  `yaml:"cleanup"`
}

type LogConfig struct {
	Level                     string `yaml:"level"`
	AuditRetentionDays        int    `yaml:"audit_retention_days"`
	SecurityLogPath           string `yaml:"security_log_path"`
	SecurityLogRetentionDays  int    `yaml:"security_log_retention_days"`
}

type SpeedConfig struct {
	ExternalUploadMbps   int `yaml:"external_upload_mbps"`
	ExternalDownloadMbps int `yaml:"external_download_mbps"`
	BurstMB              int `yaml:"burst_mb"`
}

type QuotaConfig struct {
	DefaultStorageQuotaBytes      int64 `yaml:"default_storage_quota_bytes"`
	DefaultMonthlyUploadBytes     int64 `yaml:"default_monthly_upload_bytes"`
	DefaultMonthlyDownloadBytes   int64 `yaml:"default_monthly_download_bytes"`
	DefaultMaxFileSizeBytes       int64 `yaml:"default_max_file_size_bytes"`
	DefaultMaxConcurrentUploads   int   `yaml:"default_max_concurrent_uploads"`
}

type TrafficConfig struct {
	UploadWastedAllowance   float64 `yaml:"upload_wasted_allowance"`
	DownloadWastedAllowance float64 `yaml:"download_wasted_allowance"`
	GraceFactor             float64 `yaml:"grace_factor"`
}

type RateLimitConfig struct {
	AdminLoginFailedPer15min    int `yaml:"admin_login_failed_per_15min"`
	AdminTotpFailedPer15min     int `yaml:"admin_totp_failed_per_15min"`
	InviteFailedPer15min        int `yaml:"invite_failed_per_15min"`
	UploadCreatePerHourPerson   int `yaml:"upload_create_per_hour_person"`
	ChunkPerMinutePerson        int `yaml:"chunk_per_minute_person"`
	DownloadPerMinutePerson     int `yaml:"download_per_minute_person"`
	ConcurrentDownloadsPerPerson int `yaml:"concurrent_downloads_per_person"`
	ZipPerHourPerson            int `yaml:"zip_per_hour_person"`
}

type ZipConfig struct {
	MaxFiles      int   `yaml:"max_files"`
	MaxTotalBytes int64 `yaml:"max_total_bytes"`
	StoreMode     bool  `yaml:"store_mode"`
}

type SessionConfig struct {
	UserIdleDays       int `yaml:"user_idle_days"`
	UserAbsoluteDays   int `yaml:"user_absolute_days"`
	AdminIdleHours     int `yaml:"admin_idle_hours"`
	AdminAbsoluteDays  int `yaml:"admin_absolute_days"`
}

type InviteConfig struct {
	DefaultExpiresHours     int `yaml:"default_expires_hours"`
	DefaultMaxActivations   int `yaml:"default_max_activations"`
}

type DiskConfig struct {
	MinFreeSpaceGB   int   `yaml:"min_free_space_gb"`
	CriticalFreeSpaceGB int `yaml:"critical_free_space_gb"`
	MinFreeInodes    int64 `yaml:"min_free_inodes"`
}

type UploadConfig struct {
	ChunkSizeBytes          int64   `yaml:"chunk_size_bytes"`
	MinReservationTTLHours  int     `yaml:"min_reservation_ttl_hours"`
	MaxReservationTTLHours  int     `yaml:"max_reservation_ttl_hours"`
	TTLFactor               float64 `yaml:"ttl_factor"`
	DefaultEstimatedSpeedBPS int64  `yaml:"default_estimated_speed_bps"`
}

type SecurityConfig struct {
	QuarantineSuspicious bool   `yaml:"quarantine_suspicious"`
	SessionSecretEnv     string `yaml:"session_secret_env"`
	IPHashSaltEnv        string `yaml:"ip_hash_salt_env"`
}

type CleanupConfig struct {
	RunIntervalSeconds int  `yaml:"run_interval_seconds"`
	BackupDaily        bool `yaml:"backup_daily"`
	BackupKeepDays     int  `yaml:"backup_keep_days"`
}

var DefaultConfig = Config{
	Listen:    "127.0.0.1:8090",
	BaseURL:   "https://files.example.duckdns.org",
	DataDir:   "/srv/media/fileshare/data",
	TmpDir:    "/srv/media/fileshare/tmp",
	DBPath:    "/srv/media/fileshare/db/homeshare.db",
	BackupDir: "/home/fileshare-backup",
	LocalCIDRs: []string{"192.168.32.0/24", "127.0.0.0/8"},
	Log: LogConfig{
		Level:                     "info",
		AuditRetentionDays:        180,
		SecurityLogPath:           "/var/log/homeshare/security.log",
		SecurityLogRetentionDays:  7,
	},
	SpeedLimits: SpeedConfig{
		ExternalUploadMbps:   250,
		ExternalDownloadMbps: 250,
		BurstMB:              16,
	},
	Quotas: QuotaConfig{
		DefaultStorageQuotaBytes:    100 * 1024 * 1024 * 1024,
		DefaultMonthlyUploadBytes:   200 * 1024 * 1024 * 1024,
		DefaultMonthlyDownloadBytes: 300 * 1024 * 1024 * 1024,
		DefaultMaxFileSizeBytes:     50 * 1024 * 1024 * 1024,
		DefaultMaxConcurrentUploads: 1,
	},
	Traffic: TrafficConfig{
		UploadWastedAllowance:   0.5,
		DownloadWastedAllowance: 1.0,
		GraceFactor:             0.5,
	},
	RateLimits: RateLimitConfig{
		AdminLoginFailedPer15min:    5,
		AdminTotpFailedPer15min:     5,
		InviteFailedPer15min:        5,
		UploadCreatePerHourPerson:   10,
		ChunkPerMinutePerson:        120,
		DownloadPerMinutePerson:     120,
		ConcurrentDownloadsPerPerson: 2,
		ZipPerHourPerson:            5,
	},
	SuspiciousExtensions: []string{
		"exe", "msi", "msp", "bat", "cmd", "com", "scr", "vbs", "vbe",
		"js", "jse", "ws", "wsf", "wsh", "ps1", "psm1", "sh", "bash",
		"dll", "ocx", "jar", "apk", "hta", "cpl",
	},
	Zip: ZipConfig{
		MaxFiles:      100,
		MaxTotalBytes: 50 * 1024 * 1024 * 1024,
		StoreMode:     true,
	},
	Sessions: SessionConfig{
		UserIdleDays:      30,
		UserAbsoluteDays:  90,
		AdminIdleHours:    12,
		AdminAbsoluteDays: 7,
	},
	Invite: InviteConfig{
		DefaultExpiresHours:   24,
		DefaultMaxActivations: 1,
	},
	Disk: DiskConfig{
		MinFreeSpaceGB:      40,
		CriticalFreeSpaceGB: 20,
		MinFreeInodes:       100000,
	},
	Upload: UploadConfig{
		ChunkSizeBytes:         32 * 1024 * 1024,
		MinReservationTTLHours: 1,
		MaxReservationTTLHours: 72,
		TTLFactor:              2.0,
		DefaultEstimatedSpeedBPS: 2 * 1024 * 1024,
	},
	Security: SecurityConfig{
		QuarantineSuspicious: true,
		SessionSecretEnv:     "HOMESHARE_SESSION_SECRET",
		IPHashSaltEnv:        "HOMESHARE_IP_HASH_SALT",
	},
	Cleanup: CleanupConfig{
		RunIntervalSeconds: 60,
		BackupDaily:        true,
		BackupKeepDays:     14,
	},
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func GenerateSecrets() (sessionSecret, ipHashSalt string, err error) {
	sessionBytes := make([]byte, 32)
	saltBytes := make([]byte, 32)

	if _, err := rand.Read(sessionBytes); err != nil {
		return "", "", err
	}
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(sessionBytes), hex.EncodeToString(saltBytes), nil
}

func GetSecret(envVar string) (string, error) {
	if val := os.Getenv(envVar); val != "" {
		return val, nil
	}

	secret, salt, err := GenerateSecrets()
	if err != nil {
		return "", err
	}

	if err := os.Setenv(envVar, secret); err != nil {
		return "", err
	}

	if envVar == "HOMESHARE_IP_HASH_SALT" {
		return salt, nil
	}
	return secret, nil
}

func (c *Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen address required")
	}
	if c.DataDir == "" {
		return fmt.Errorf("data_dir required")
	}
	if c.TmpDir == "" {
		return fmt.Errorf("tmp_dir required")
	}
	if c.DBPath == "" {
		return fmt.Errorf("db_path required")
	}
	if c.SpeedLimits.ExternalUploadMbps < 10 || c.SpeedLimits.ExternalUploadMbps > 1000 {
		return fmt.Errorf("external_upload_mbps must be 10-1000")
	}
	if c.SpeedLimits.ExternalDownloadMbps < 10 || c.SpeedLimits.ExternalDownloadMbps > 1000 {
		return fmt.Errorf("external_download_mbps must be 10-1000")
	}
	if c.SpeedLimits.BurstMB < 1 || c.SpeedLimits.BurstMB > 128 {
		return fmt.Errorf("burst_mb must be 1-128")
	}
	return nil
}

func (c *Config) ExternalUploadBytesPerSec() int64 {
	return int64(c.SpeedLimits.ExternalUploadMbps) * 1_000_000 / 8
}

func (c *Config) ExternalDownloadBytesPerSec() int64 {
	return int64(c.SpeedLimits.ExternalDownloadMbps) * 1_000_000 / 8
}

func (c *Config) BurstBytes() int64 {
	return int64(c.SpeedLimits.BurstMB) * 1024 * 1024
}

func (c *Config) MinFreeSpaceBytes() int64 {
	return int64(c.Disk.MinFreeSpaceGB) * 1024 * 1024 * 1024
}

func (c *Config) CriticalFreeSpaceBytes() int64 {
	return int64(c.Disk.CriticalFreeSpaceGB) * 1024 * 1024 * 1024
}

func (c *Config) UserIdleTimeout() time.Duration {
	return time.Duration(c.Sessions.UserIdleDays) * 24 * time.Hour
}

func (c *Config) UserAbsoluteTimeout() time.Duration {
	return time.Duration(c.Sessions.UserAbsoluteDays) * 24 * time.Hour
}

func (c *Config) AdminIdleTimeout() time.Duration {
	return time.Duration(c.Sessions.AdminIdleHours) * time.Hour
}

func (c *Config) AdminAbsoluteTimeout() time.Duration {
	return time.Duration(c.Sessions.AdminAbsoluteDays) * 24 * time.Hour
}
