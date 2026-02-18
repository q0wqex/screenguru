package main

import (
	"os"
	"strconv"
	"time"
)

// Server configuration
const (
	ServerAddr = "0.0.0.0:8000"
)

// File system configuration
const (
	DataPath       = "/data"
	TemplatesPath  = "templates"
	StaticPath     = "templates/static"
	SecretFilePath = DataPath + "/.secret"
	ChangelogPath  = "../changelog.md"
	ChangelogURL   = "https://raw.githubusercontent.com/q0wqex/screenguru/main/changelog.md"

	DefaultFilePerm = 0755
)

var (
	MaxFileSize = int64(10 * 1024 * 1024) // 10MB default
)

// MIME types and extensions
var (
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	ImageExtensions = map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"image/gif":  "gif",
		"image/webp": "webp",
	}

	// AppSecret is used to sign cookies. It's loaded on startup.
	AppSecret []byte
)

// Session configuration
const (
	SessionCookieName = "session_id"
	SessionMaxAge     = 86400 * 30 // 30 days
)

// Cleanup configuration
var (
	CleanupDuration = 720 * time.Hour // 30 days default
	CleanupInterval = 24 * time.Hour  // 24 hours default
)

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	if maxSizeStr := os.Getenv("MAX_FILE_SIZE_MB"); maxSizeStr != "" {
		if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			MaxFileSize = size * 1024 * 1024
		}
	}

	if cleanupHoursStr := os.Getenv("CLEANUP_DURATION_HOURS"); cleanupHoursStr != "" {
		if hours, err := strconv.Atoi(cleanupHoursStr); err == nil {
			CleanupDuration = time.Duration(hours) * time.Hour
		}
	}
}
