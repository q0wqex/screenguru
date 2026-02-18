package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Path builders for data directories
func userPath(userID string) string           { return filepath.Join(DataPath, userID) }
func albumPath(userID, albumID string) string { return filepath.Join(DataPath, userID, albumID) }
func imagePath(userID, albumID, filename string) string {
	return filepath.Join(DataPath, userID, albumID, filename)
}

// Глобальная переменная для хранения общего количества изображений
var TotalImageCount int

// ImageInfo хранит информацию об изображении
type ImageInfo struct {
	Filename string
	Path     string
	Size     int64
	UserID   string
	AlbumID  string
}

// AlbumInfo хранит информацию об альбоме
type AlbumInfo struct {
	ID         string
	Name       string
	ImageCount int
	CreatedAt  time.Time
}

// saveImage сохраняет загруженное изображение
func saveImage(file multipart.File, header *multipart.FileHeader, userID, albumID string) (*ImageInfo, error) {
	// Проверка размера файла
	if header.Size > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes", header.Size)
	}

	// Валидация типа изображения
	extension, valid := validateImageType(file)
	if !valid {
		return nil, fmt.Errorf("invalid image type")
	}

	// Создание директории для альбома
	albumPath := albumPath(userID, albumID)
	if err := EnsureDir(albumPath); err != nil {
		return nil, err
	}

	// Генерация уникального имени файла
	filename := generateUniqueFilename(extension)
	filePath := albumPath + "/" + filename

	// Создание файла
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// Копирование содержимого
	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	// Получение информации о файле
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// Увеличиваем глобальный счетчик изображений
	TotalImageCount++

	return &ImageInfo{
		Filename: filename,
		Path:     filePath,
		Size:     stat.Size(),
		UserID:   userID,
		AlbumID:  albumID,
	}, nil
}

// validateImageType проверяет тип изображения
func validateImageType(file multipart.File) (string, bool) {
	// Чтение заголовка файла
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", false
	}

	// Восстановление указателя
	file.Seek(0, 0)

	// Определение MIME типа
	contentType := http.DetectContentType(buffer)

	// Проверка разрешенных типов
	if !AllowedImageTypes[contentType] {
		return "", false
	}

	// Возвращаем соответствующее расширение
	if ext, exists := ImageExtensions[contentType]; exists {
		return ext, true
	}

	return "", false
}

// generateUniqueFilename генерирует уникальное имя файла
func generateUniqueFilename(extension string) string {
	ext := strings.ToLower(extension)
	if ext == "" {
		ext = ".webp" // расширение по умолчанию
	} else if !strings.HasPrefix(ext, ".") {
		ext = "." + ext // добавляем точку если её нет
	}

	randomID := RandomID()
	return randomID + ext
}

// getUserImages возвращает список изображений пользователя
func getUserImages(userID, albumID string) ([]ImageInfo, error) {
	dirPath := albumPath(userID, albumID)

	// Проверка существования директории
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return []ImageInfo{}, nil
	}

	// Чтение содержимого директории
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var images []ImageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !IsImageFile(filename) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		images = append(images, ImageInfo{
			Filename: filename,
			Path:     filepath.Join(dirPath, filename),
			Size:     info.Size(),
			UserID:   userID,
			AlbumID:  albumID,
		})
	}

	// Сортировка изображений по времени модификации (старые сверху, новые снизу)
	sort.Slice(images, func(i, j int) bool {
		infoI, errI := os.Stat(images[i].Path)
		infoJ, errJ := os.Stat(images[j].Path)

		if errI != nil || errJ != nil {
			return false // если не можем получить статус файла, не меняем порядок
		}

		return infoI.ModTime().Before(infoJ.ModTime())
	})

	return images, nil
}

// getSessionID получает или генерирует ID сессии пользователя с проверкой подписи
func getSessionID(w http.ResponseWriter, r *http.Request) string {
	// Проверка наличия cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		parts := strings.Split(cookie.Value, ":")
		if len(parts) == 2 {
			userID := parts[0]
			signature := parts[1]

			// Проверка подписи
			if VerifyData(userID, signature) {
				logger.Debug(fmt.Sprintf("getSessionID: signature verified, userID=%s", userID))
				return userID
			}
			logger.Error(fmt.Sprintf("getSessionID: INVALID signature for userID=%s", userID))
		} else if len(parts) == 1 {
			// МИГРАЦИЯ: Если кука без подписи, проверяем существует ли такой пользователь
			oldUserID := parts[0]
			userDir := userPath(oldUserID)
			if info, err := os.Stat(userDir); err == nil && info.IsDir() {
				logger.Info(fmt.Sprintf("getSessionID: Migrating old user %s to signed session", oldUserID))

				// Подписываем старый ID и обновляем куку
				signature := SignData(oldUserID)
				signedValue := fmt.Sprintf("%s:%s", oldUserID, signature)

				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookieName,
					Value:    signedValue,
					Path:     "/",
					MaxAge:   SessionMaxAge,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				return oldUserID
			}
		}
	}

	// Генерация нового ID сессии
	sessionID := RandomID()
	signature := SignData(sessionID)
	signedValue := fmt.Sprintf("%s:%s", sessionID, signature)

	logger.Debug(fmt.Sprintf("getSessionID: creating new signed session, sessionID=%s", sessionID))

	// Установка cookie с подписью
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    signedValue,
		Path:     "/",
		MaxAge:   SessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return sessionID
}

// getUserAlbums возвращает список альбомов пользователя
func getUserAlbums(userID string) ([]AlbumInfo, error) {
	userDir := userPath(userID)
	logger.Debug(fmt.Sprintf("getUserAlbums: userDir=%s", userDir))

	// Проверка существования директории
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		logger.Debug("getUserAlbums: user dir does not exist")
		return []AlbumInfo{}, nil
	}

	// Чтение содержимого директории
	entries, err := os.ReadDir(userDir)
	if err != nil {
		logger.Debug(fmt.Sprintf("getUserAlbums: error reading dir: %v", err))
		return nil, err
	}
	logger.Debug(fmt.Sprintf("getUserAlbums: found %d entries", len(entries)))

	var albums []AlbumInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		albumID := entry.Name()
		albumDir := filepath.Join(userDir, albumID)

		// Получение информации о директории
		dirInfo, err := os.Stat(albumDir)
		var createdAt time.Time
		if err == nil {
			createdAt = dirInfo.ModTime()
		}

		// Подсчет количества изображений
		imageCount := countImagesInDir(albumDir)

		// Добавление альбома в список
		albums = append(albums, AlbumInfo{
			ID:         albumID,
			Name:       albumID,
			ImageCount: imageCount,
			CreatedAt:  createdAt,
		})
	}

	// Сортировка альбомов по дате создания (новые сверху)
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].CreatedAt.Equal(albums[j].CreatedAt) {
			return albums[i].ID > albums[j].ID
		}
		return albums[i].CreatedAt.After(albums[j].CreatedAt)
	})

	return albums, nil
}

// countImagesInDir подсчитывает количество изображений в директории
func countImagesInDir(dirPath string) int {
	count := 0
	processDir(dirPath, func(entry os.DirEntry) bool {
		return !entry.IsDir() && IsImageFile(entry.Name())
	}, func(path string, info os.FileInfo) error {
		count++
		return nil
	})
	return count
}

// createAlbum создает новый альбом для пользователя
func createAlbum(userID string) (string, error) {
	albumID := RandomID()

	// Создание директории для альбома
	albumDir := albumPath(userID, albumID)
	logger.Debug(fmt.Sprintf("createAlbum: creating albumDir=%s", albumDir))
	if err := EnsureDir(albumDir); err != nil {
		return "", err
	}
	logger.Debug(fmt.Sprintf("createAlbum: album created, albumID=%s", albumID))

	return albumID, nil
}

// deleteImage удаляет изображение
func deleteImage(userID, albumID, filename string) error {
	filePath := imagePath(userID, albumID, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("image not found")
	}

	err := os.Remove(filePath)
	if err == nil {
		// Уменьшаем глобальный счетчик изображений
		TotalImageCount--
	}
	return err
}

// deleteAlbum удаляет альбом со всеми изображениями
func deleteAlbum(userID, albumID string) error {
	albumDir := albumPath(userID, albumID)

	if _, err := os.Stat(albumDir); os.IsNotExist(err) {
		return fmt.Errorf("album not found")
	}

	// Подсчитываем количество изображений в альбоме перед удалением
	imageCount := countImagesInDir(albumDir)

	err := os.RemoveAll(albumDir)
	if err == nil {
		// Уменьшаем глобальный счетчик изображений на количество удаленных изображений
		TotalImageCount -= imageCount
	}
	return err
}

// deleteUser удаляет все данные пользователя
func deleteUser(userID string) error {
	userDir := userPath(userID)

	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return fmt.Errorf("user directory not found")
	}

	// Подсчитываем количество изображений в пользовательской директории перед удалением
	totalImages := 0
	// Рекурсивный обход всей директории пользователя
	err := filepath.Walk(userDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Пропускаем ошибки доступа к файлам
			return nil
		}

		// Пропускаем директории
		if !info.IsDir() {
			// Проверяем, является ли файл изображением
			if IsImageFile(info.Name()) {
				totalImages++
			}
		}

		return nil
	})

	errRemove := os.RemoveAll(userDir)
	if errRemove == nil && err == nil {
		// Уменьшаем глобальный счетчик изображений на количество удаленных изображений
		TotalImageCount -= totalImages
	}
	return errRemove
}

// countAllFilesInDataPath подсчитывает количество всех файлов в директории data при запуске приложения
func countAllFilesInDataPath() int {
	count := 0

	// Рекурсивный обход всей директории DataPath
	err := filepath.Walk(DataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Пропускаем ошибки доступа к файлам
			return nil
		}

		// Пропускаем директории
		if !info.IsDir() {
			count++
		}

		return nil
	})

	if err != nil {
		logger.Error(fmt.Sprintf("countAllFilesInDataPath: error walking data directory: %v", err))
	}

	return count
}

// loadOrGenerateSecret загружает секрет из файла или генерирует новый
func loadOrGenerateSecret() error {
	// Если секрет уже загружен (например, через окружение, хотя сейчас мы через файл)
	if len(AppSecret) > 0 {
		return nil
	}

	// Пробуем прочитать из файла
	data, err := os.ReadFile(SecretFilePath)
	if err == nil && len(data) >= 32 {
		AppSecret = data
		logger.Info("App secret loaded from file")
		return nil
	}

	// Генерируем новый секрет
	secret := make([]byte, 32)
	// Используем crypto/rand напрямую здесь для безопасности
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("failed to generate random secret: %v", err)
	}

	// Сохраняем в файл
	if err := os.WriteFile(SecretFilePath, secret, 0600); err != nil {
		logger.Error(fmt.Sprintf("Failed to save secret to file: %v. Sessions will not persist across restarts.", err))
	} else {
		logger.Info("New app secret generated and saved")
	}

	AppSecret = secret
	return nil
}
