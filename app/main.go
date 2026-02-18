package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Template cache for improved performance
var templates *template.Template

func main() {
	// Инициализация приложения
	if err := initializeApp(); err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	// Создание HTTP роутера
	mux := setupRoutes()

	// Запуск cleanup worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startCleanupWorker(ctx)

	// Настройка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("Server starting on %s\n", ServerAddr)
		serverErr <- http.ListenAndServe(ServerAddr, mux)
	}()

	// Ожидание сигнала или ошибки
	select {
	case <-sigChan:
		fmt.Println("Received shutdown signal")
	case err := <-serverErr:
		fmt.Printf("Server error: %v\n", err)
	}

	fmt.Println("Shutting down gracefully...")
	cancel()
}

// initializeApp инициализирует приложение
func initializeApp() error {
	// Загружаем конфигурацию из переменных окружения
	LoadConfig()

	// Создание директории для хранения данных
	if err := EnsureDir(DataPath); err != nil {
		return err
	}

	// Подсчет общего количества изображений при запуске приложения
	TotalImageCount = countAllFilesInDataPath()
	logger.Info(fmt.Sprintf("Total images on startup: %d", TotalImageCount))

	// Инициализация секретного ключа для подписи куки
	if err := loadOrGenerateSecret(); err != nil {
		return fmt.Errorf("failed to initialize secret: %w", err)
	}

	// Проверка доступности директории шаблонов
	if err := checkTemplates(); err != nil {
		return err
	}

	return nil
}

// checkTemplates загружает и кеширует шаблоны
func checkTemplates() error {
	tmpl, err := template.ParseGlob(TemplatesPath + "/*.html")
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}
	templates = tmpl
	return nil
}

// setupRoutes настраивает HTTP роуты
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Статические файлы
	mux.HandleFunc("/static/", handleStaticFiles)
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, StaticPath+"/robots.txt")
	})
	mux.HandleFunc("/sitemap.xml", sitemapHandler)

	// API endpoints
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/create-album", createAlbumHandler)
	mux.HandleFunc("/delete-image", deleteImageHandler)
	mux.HandleFunc("/delete-album", deleteAlbumHandler)
	mux.HandleFunc("/delete-user", deleteUserHandler)
	mux.HandleFunc("/changelog", changelogHandler)

	return mux
}

// handleStaticFiles обрабатывает статические файлы
func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	filePath := StaticPath + r.URL.Path[len("/static"):]

	// Определение MIME типа
	setContentType(w, filePath)

	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, filePath)
}

// setContentType устанавливает Content-Type для статических файлов
func setContentType(w http.ResponseWriter, filePath string) {
	switch GetFileExtension(filePath) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}
}
