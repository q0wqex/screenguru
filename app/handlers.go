package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// indexHandler обрабатывает главную страницу
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	logger.Debug(fmt.Sprintf("indexHandler: request received, path=%s, method=%s", r.URL.Path, r.Method))
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Если это не главная страница, обрабатываем как контент
	if r.URL.Path != "/" {
		contentHandler(w, r)
		return
	}

	// Получаем сессию пользователя
	sessionID := getSessionID(w, r)

	// Получаем список альбомов
	albums, err := getUserAlbums(sessionID)
	logger.Debug(fmt.Sprintf("getUserAlbums: sessionID=%s, albums_count=%d, err=%v", sessionID, len(albums), err))
	if err != nil {
		albums = []AlbumInfo{}
	}

	// Подготавливаем данные для шаблона
	data := struct {
		Albums          []AlbumInfo
		HasAlbums       bool
		SessionID       string
		TotalImageCount int
	}{
		Albums:          albums,
		HasAlbums:       len(albums) > 0,
		SessionID:       sessionID,
		TotalImageCount: TotalImageCount,
	}

	// Отображаем страницу
	if err := renderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// uploadHandler обрабатывает загрузку изображений
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug(fmt.Sprintf("uploadHandler: request received, method=%s", r.Method))
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getSessionID(w, r)
	logger.Debug(fmt.Sprintf("uploadHandler: sessionID=%s", sessionID))

	// Ограничиваем размер запроса
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Получаем ID альбома
	albumID := getAlbumID(r, sessionID)

	// Проверяем файлы
	files := getUploadFiles(r)
	if len(files) == 0 {
		http.Error(w, "No files selected", http.StatusBadRequest)
		return
	}

	// Обрабатываем файлы
	if err := processUpload(files, sessionID, albumID); err != nil {
		http.Error(w, fmt.Sprintf("Upload failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Проверяем, является ли запрос XHR (технический/фоновый)
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.Header.Get("Accept") == "application/json" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Перенаправляем на альбом
	http.Redirect(w, r, "/"+sessionID+"/"+albumID, http.StatusSeeOther)
}

// contentHandler обрабатывает отдачу изображений или страницы альбома
func contentHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.SplitN(path, "/", 3)

	switch len(parts) {
	case 2:
		// Страница альбома
		handleAlbumPage(w, r, parts[0], parts[1])
	case 3:
		// Файл изображения
		handleImageFile(w, r, parts[0], parts[1], parts[2])
	default:
		http.NotFound(w, r)
	}
}

// handleAlbumPage обрабатывает страницу альбома
func handleAlbumPage(w http.ResponseWriter, r *http.Request, sessionID, albumID string) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	logger.Debug(fmt.Sprintf("handleAlbumPage: sessionID=%s, albumID=%s", sessionID, albumID))
	currentSessionID := getSessionID(w, r)
	isOwner := currentSessionID == sessionID

	images, _ := getUserImages(sessionID, albumID)
	logger.Debug(fmt.Sprintf("handleAlbumPage: images_count=%d", len(images)))

	data := struct {
		Images          []ImageInfo
		HasImages       bool
		SessionID       string
		OwnerSessionID  string
		AlbumID         string
		IsOwner         bool
		TotalImageCount int
	}{
		Images:          images,
		HasImages:       len(images) > 0,
		SessionID:       currentSessionID,
		OwnerSessionID:  sessionID,
		AlbumID:         albumID,
		IsOwner:         isOwner,
		TotalImageCount: TotalImageCount,
	}

	if err := renderTemplate(w, "album.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleImageFile обрабатывает отдачу файла изображения
func handleImageFile(w http.ResponseWriter, r *http.Request, sessionID, albumID, filename string) {
	filePath := imagePath(sessionID, albumID, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

// deleteImageHandler обрабатывает удаление изображения
func deleteImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getSessionID(w, r)
	albumID := r.FormValue("album_id")
	filename := r.FormValue("filename")

	if albumID == "" || filename == "" {
		http.Error(w, "album_id and filename required", http.StatusBadRequest)
		return
	}

	if err := deleteImage(sessionID, albumID, filename); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting image: %v", err), http.StatusInternalServerError)
		return
	}

	SuccessResponse(w, map[string]string{"message": "Image deleted successfully"})
}

// deleteAlbumHandler обрабатывает удаление альбома
func deleteAlbumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getSessionID(w, r)
	albumID := r.FormValue("album_id")

	if albumID == "" {
		http.Error(w, "album_id required", http.StatusBadRequest)
		return
	}

	if err := deleteAlbum(sessionID, albumID); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting album: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// deleteUserHandler обрабатывает удаление профиля пользователя
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getSessionID(w, r)

	if err := deleteUser(sessionID); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting user data: %v", err), http.StatusInternalServerError)
		return
	}

	// Очищаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	SuccessResponse(w, map[string]string{"message": "Profile deleted successfully"})
}

// createAlbumHandler создает новый альбом и возвращает его ID
func createAlbumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := getSessionID(w, r)

	albumID, err := createAlbum(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating album: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"album_id": "%s", "session_id": "%s"}`, albumID, sessionID)
}

// changelogCache хранит содержимое ченджлога в памяти
var (
	changelogCache      string
	changelogCacheMutex sync.RWMutex
	lastFetchTime       time.Time
)

// changelogHandler возвращает содержимое ченджлога
func changelogHandler(w http.ResponseWriter, r *http.Request) {
	changelogCacheMutex.RLock()
	// Кэшируем на 1 час
	if changelogCache != "" && time.Since(lastFetchTime) < time.Hour {
		content := changelogCache
		changelogCacheMutex.RUnlock()
		SuccessResponse(w, map[string]string{"content": content})
		return
	}
	changelogCacheMutex.RUnlock()

	var content string
	// Пробуем прочитать локально (в разных возможных путях)
	paths := []string{ChangelogPath, "./changelog.md", "changelog.md"}
	for _, p := range paths {
		bytes, err := os.ReadFile(p)
		if err == nil {
			content = string(bytes)
			break
		}
	}

	if content == "" {
		// Если локально нет, тянем с GitHub
		logger.Info("Local changelog not found, fetching from GitHub...")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(ChangelogURL)
		if err != nil {
			http.Error(w, "Changelog not found", http.StatusNotFound)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Changelog not found on GitHub", http.StatusNotFound)
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading changelog", http.StatusInternalServerError)
			return
		}
		content = string(bodyBytes)
	}

	// Обновляем кэш
	changelogCacheMutex.Lock()
	changelogCache = content
	lastFetchTime = time.Now()
	changelogCacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	SuccessResponse(w, map[string]string{"content": content})
}

// Вспомогательные функции

// getAlbumID получает или создает ID альбома
func getAlbumID(r *http.Request, sessionID string) string {
	albumID := r.FormValue("album_id")
	if albumID != "" {
		return albumID
	}

	// Создаем новый альбом если не указан
	newAlbumID, err := createAlbum(sessionID)
	if err != nil {
		return ""
	}
	return newAlbumID
}

// getUploadFiles извлекает файлы из запроса
func getUploadFiles(r *http.Request) []*multipart.FileHeader {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil
	}

	files := r.MultipartForm.File["image"]
	if len(files) == 0 {
		return nil
	}

	return files
}

// processUpload обрабатывает загрузку файлов параллельно
func processUpload(files []*multipart.FileHeader, sessionID, albumID string) error {
	logger.Debug(fmt.Sprintf("processUpload: starting, files_count=%d, sessionID=%s, albumID=%s", len(files), sessionID, albumID))
	var wg sync.WaitGroup
	errs := make(chan error, len(files))

	for _, fileHeader := range files {
		wg.Add(1)
		go func(fh *multipart.FileHeader) {
			defer wg.Done()
			file, err := fh.Open()
			if err != nil {
				errs <- fmt.Errorf("error opening file %s: %v", fh.Filename, err)
				return
			}
			defer file.Close()

			_, err = saveImage(file, fh, sessionID, albumID)
			if err != nil {
				errs <- fmt.Errorf("error saving file %s: %v", fh.Filename, err)
				return
			}
		}(fileHeader)
	}

	wg.Wait()
	close(errs)

	var uploadErrors []string
	for err := range errs {
		uploadErrors = append(uploadErrors, err.Error())
	}

	if len(uploadErrors) > 0 {
		return fmt.Errorf("%v", strings.Join(uploadErrors, "; "))
	}
	return nil
}

// renderTemplate рендерит HTML шаблон из кеша
func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	return templates.ExecuteTemplate(w, name, data)
}

// sitemapHandler генерирует sitemap.xml
func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	// Всегда используем https и основной домен для sitemap
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://screengu.ru/</loc>
    <lastmod>%s</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>`, time.Now().Format("2006-01-02"))
}
