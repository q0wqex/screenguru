package main

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// startCleanupWorker запускает фоновый процесс очистки старых изображений
func startCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Graceful shutdown
		case <-ticker.C:
			performCleanup()
		}
	}
}

// performCleanup выполняет очистку
func performCleanup() {
	if err := cleanupOldImages(); err != nil {
		logger.Error("Failed to cleanup old images: " + err.Error())
	}

	if err := removeEmptyDirectories(); err != nil {
		logger.Error("Failed to remove empty directories: " + err.Error())
	}
}

// cleanupOldImages удаляет старые изображения
func cleanupOldImages() error {
	// Проверка существования директории /data
	if _, err := os.Stat(DataPath); os.IsNotExist(err) {
		return nil
	}

	// Чтение всех пользовательских директорий
	entries, err := os.ReadDir(DataPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		userDir := filepath.Join(DataPath, entry.Name())
		if err := cleanupUserImages(userDir); err != nil {
			logger.Error("Failed to cleanup user images in " + userDir + ": " + err.Error())
		}
	}

	return nil
}

// cleanupUserImages очищает старые изображения в директории пользователя
func cleanupUserImages(userDir string) error {
	return processDir(userDir, func(entry os.DirEntry) bool {
		return !entry.IsDir()
	}, func(filePath string, info os.FileInfo) error {
		if isImageOld(info.ModTime()) {
			if err := os.Remove(filePath); err != nil {
				logger.Error("Failed to remove old image " + filePath + ": " + err.Error())
			}
		}
		return nil
	})
}

// removeEmptyDirectories удаляет пустые директории
func removeEmptyDirectories() error {
	// Проверка существования директории /data
	if _, err := os.Stat(DataPath); os.IsNotExist(err) {
		return nil
	}

	// Чтение всех пользовательских директорий
	entries, err := os.ReadDir(DataPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		userDir := filepath.Join(DataPath, entry.Name())

		// Проверяем, пуста ли директория
		isEmpty, err := isDirEmpty(userDir)
		if err != nil {
			continue
		}

		if isEmpty {
			if err := os.Remove(userDir); err != nil {
				logger.Error("Failed to remove empty directory " + userDir + ": " + err.Error())
			}
		}
	}

	return nil
}

// isDirEmpty проверяет, пуста ли директория
func isDirEmpty(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}