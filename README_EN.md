# Screenguru

[![Stars](https://img.shields.io/github/stars/q0wqex/screenguru.svg?style=social)](https://github.com/q0wqex/screenguru/stargazers)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue?logo=docker)](https://github.com/q0wqex/screenguru/pkgs/container/screenguru)

[🇷🇺 Русский](README.md) | [🇺🇸 English](README_EN.md)

Lightweight self-hosted image hosting in Go. Zero-framework, flat-file storage, TTL-based cleanup.

## Quick Start

```yaml
services:
  screenguru:
    image: ghcr.io/q0wqex/screenguru:main
    ports:
      - "8000:8000"
    volumes:
      - ./data:/data
    environment:
      - MAX_FILE_SIZE_MB=10
      - CLEANUP_DURATION_HOURS=720
    restart: unless-stopped
```

## Environment Variables

- `MAX_FILE_SIZE_MB`: Upload limit (default: 10)
- `CLEANUP_DURATION_HOURS`: File TTL (default: 720)

## API

- `GET /`: Index/Album
- `POST /upload`: Upload (`files`, `album_id`)
- `POST /create-album`: Generate ID
- `POST /delete-image`: Delete (`image_id`, `album_id`)
- `POST /delete-album`: Recursive delete (`album_id`)
- `POST /delete-user`: Profile delete (session-based)

## Development

```bash
cd app && go build -ldflags="-s -w" -o screenguru && ./screenguru
```

---
*Performance and aesthetics.*
