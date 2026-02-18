# Screenguru

[![Stars](https://img.shields.io/github/stars/q0wqex/screenguru.svg?style=social)](https://github.com/q0wqex/screenguru/stargazers)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/q0wqex/screenguru/pkgs/container/screenguru)
[![License](https://img.shields.io/badge/license-AGPL--3.0-orange)](LICENSE)

[🇷🇺 Русский](README.md) | [🇺🇸 English](README_EN.md)

## Description

A lightweight and fast image hosting service written in Go. Created as a simple alternative for those seeking minimalism and speed.

> *Screenguru — when the habit of simple and beautiful hosting remains forever.*

## Features

- **🖼️ Image Support**: Upload and view JPEG, PNG, GIF, WebP.
- **📁 Albums**: Organize images into albums.
- **🧹 Auto-Cleanup**: Automatic removal of old files (default 30 days).
- **🚀 Ultra Fast**: Minimal dependencies, using cached templates.
- **🐳 Docker Ready**: Full support for containerization and one-command deployment.
- **📱 Responsive UI**: Adaptive interface for mobile and desktop.

## API

The server runs on port `8000` by default.

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/` | GET | Main page / Album view |
| `/upload` | POST | Upload an image |
| `/create-album` | POST | Create a new album |
| `/delete-image` | POST | Delete a specific image |
| `/delete-album` | POST | Delete an entire album |
| `/delete-user` | POST | Delete user and all their data |
| `/changelog` | GET | View change history |

## Configuration

Core parameters are defined in `app/config.go`:

| Variable | Default Value | Description |
|----------|---------------|-------------|
| `MaxFileSize` | `10 * 1024 * 1024` (10MB) | Maximum upload file size |
| `CleanupDuration` | `30 days` | File storage duration before deletion |
| `CleanupInterval` | `24 hours` | Frequency of old file checks |
| `DataPath` | `/data` | Path to image storage directory |
| `ServerAddr` | `0.0.0.0:8000` | Server address and port |

## Setup Instructions

### 1. Quick Start (Docker)

Start the service with a single command (creates data folder, downloads config, and starts container):

```bash
mkdir -p screenguru/data && cd screenguru && curl -O https://raw.githubusercontent.com/q0wqex/screenguru/main/docker-compose.yml && docker-compose up -d
```

### 2. Manual Installation

If you want to build from source:

1. Clone the repository:
   ```bash
   git clone https://github.com/q0wqex/screenguru.git && cd screenguru
   ```
2. Build and run:
   ```bash
   cd app && go build -o screenguru && ./screenguru
   ```

## Reverse Proxy & SSL Configuration (Caddy)

Caddy now automatically handles two domains: `screengu.ru` and `dev.screengu.ru`.

1. **Caddyfile**: Located in the project root. Proxying is configured as follows:
   - `screengu.ru` -> `localhost:8000`
   - `dev.screengu.ru` -> `localhost:8888`
2. **Automatic SSL**: Caddy will automatically obtain and renew certificates for both domains.
3. **Security**: The application is bound to `127.0.0.1`.

4. **Data**: Certificates are stored in the `caddy_data` folder.

The Caddy upload limit is set to 10MB (`request_body_limit`).




## Project Structure

- `/app` — Go server source code.
- `/app/templates` — HTML templates and static files (JS/CSS).
- `/data` — Image storage (created automatically).
- `docker-compose.yml` — Docker deployment file.
- `changelog.md` — Project history.

## Acknowledgments

- **[Remnawave](https://remna.st/)** — for the great visual style and design that became the foundation for this project's look.
