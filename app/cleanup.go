package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// startCleanupWorker запускает фоновый процесс очистки старых изображений
func startCleanupWorker(ctx context.Context) {
	// Первая очистка сразу при запуске
	performCleanup()

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
	logger.Info("Starting background cleanup...")
	if err := cleanupRecursive(DataPath); err != nil {
		logger.Error("Cleanup failed: " + err.Error())
	} else {
		logger.Info("Cleanup completed successfully")
	}
}

// cleanupRecursive рекурсивно удаляет старые файлы и пустые директории
func cleanupRecursive(root string) error {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	var dirs []string
	deletedFiles := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Пропускаем файлы с ошибками доступа
		}

		// Исключаем скрытые или системные файлы (начинающиеся с точки)
		// Но не саму корневую директорию /data
		name := info.Name()
		if strings.HasPrefix(name, ".") && path != root {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if path != root {
				dirs = append(dirs, path)
			}
			return nil
		}

		// Проверяем срок жизни файла используя вспомогательную функцию из utils.go
		if isImageOld(info.ModTime()) {
			if err := os.Remove(path); err != nil {
				logger.Error("Failed to remove old file " + path + ": " + err.Error())
			} else {
				deletedFiles++
				// Обновляем глобальный счетчик, если это изображение
				if IsImageFile(name) {
					TotalImageCount--
				}
				logger.Debug("Removed old file: " + path)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if deletedFiles > 0 {
		logger.Info(fmt.Sprintf("Cleanup: deleted %d expired files", deletedFiles))
	}

	// Удаление пустых директорий "снизу-вверх"
	// Используем накопленный список директорий и проходим его с конца
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		isEmpty, err := isDirEmpty(dir)
		if err != nil {
			continue
		}

		if isEmpty {
			if err := os.Remove(dir); err != nil {
				logger.Error("Failed to remove empty directory " + dir + ": " + err.Error())
			} else {
				logger.Debug("Removed empty directory: " + dir)
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