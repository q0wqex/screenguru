package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// dirEntryFilter определяет условие фильтрации для записи директории
type dirEntryFilter func(entry os.DirEntry) bool

// processDir обрабатывает все записи в директории с помощью фильтра
func processDir(dirPath string, filter dirEntryFilter, process func(path string, info os.FileInfo) error) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if filter != nil && !filter(entry) {
			continue
		}

		entryPath := filepath.Join(dirPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if err := process(entryPath, info); err != nil {
			return err
		}
	}
	return nil
}

// isImageOld проверяет, является ли изображение старым
func isImageOld(modTime time.Time) bool {
	return time.Since(modTime) > CleanupDuration
}

// Logger - простая структура для логирования
type Logger struct {
	debug bool
}

func NewLogger(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (l *Logger) Debug(msg string) {
	if l.debug {
		log.Printf("[DEBUG] %s", msg)
	}
}

func (l *Logger) Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

func (l *Logger) Error(msg string) {
	log.Printf("[ERROR] %s", msg)
}

// Global logger instance
var logger = NewLogger(os.Getenv("DEBUG") == "true")

// ErrorResponse отправляет JSON ответ с ошибкой
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s"}`, strings.ReplaceAll(message, `"`, `\\"`))
}

// SuccessResponse отправляет JSON ответ с успешным результатом
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, `{"success": true, "data": %s}`, string(jsonData))
}

// ValidatePath проверяет безопасность пути
func ValidatePath(path string) bool {
	// Проверяем на попытки выйти за пределы директории
	cleanPath := filepath.Clean(path)
	return !strings.Contains(cleanPath, "..") && !strings.HasPrefix(cleanPath, "/")
}

// EnsureDir создает директорию если она не существует
func EnsureDir(path string) error {
	return os.MkdirAll(path, DefaultFilePerm)
}

// GetFileExtension возвращает расширение файла в нижнем регистре
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// IsImageFile проверяет является ли файл изображением
var ValidImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func IsImageFile(filename string) bool {
	ext := GetFileExtension(filename)
	return ValidImageExtensions[ext]
}

// RandomID генерирует случайный ID
func RandomID() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback на timestamp
		return fmt.Sprintf("%05x", time.Now().UnixNano()%1048576)
	}
	return hex.EncodeToString(bytes)[:5]
}

// SignData генерирует HMAC-SHA256 подпись для данных
func SignData(data string) string {
	h := hmac.New(sha256.New, AppSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyData проверяет HMAC-SHA256 подпись
func VerifyData(data, signature string) bool {
	if signature == "" {
		return false
	}
	expectedSignature := SignData(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
