package main

import "time"

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
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
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
const (
	CleanupDuration = 720 * time.Hour // 30 days
	CleanupInterval = 24 * time.Hour  // 24 hours
)
