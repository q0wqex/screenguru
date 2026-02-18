---
trigger: always_on
---

# Antigravity's Project State & Rules

This file defines the persistent state, personality, and rules for the Antigravity AI assistant working on the **Скрингуру** project.

## 🧠 Project Context
*   **Name:** Скрингуру
*   **Nature:** Lightweight, fast, and visually stunning image hosting/album service (screengu.ru).
*   **Tech Stack:** Go (Backend) + Vanilla HTML/CSS/JS (Frontend). No heavy frameworks unless strictly necessary.
*   **Design Philosophy:** Glassmorphism, premium aesthetics, smooth animations (e.g., Cat Loader).

## 🛠 Coding Standards
1.  **UI First:** Every UI change must feel premium. Use curated HSL colors, Google Fonts (Montserrat, Unbounded), and subtle gradients.
2.  **No Placeholders:** Never use placeholder images. Generate real assets using tools.
3.  **Modern CSS:** Use CSS variables, Flexbox/Grid, and `backdrop-filter` for the Glassmorphism effect.
4.  **Backend Simplicity:** Keep Go code idiomatic, performant, and well-structured.

## 🔄 Workflow Rules (Global)
1.  **Branching:** Always work in the `dev` branch. `main` is ONLY for stable, released code.
2.  **Changelog is Law:** Every feature or fix must be documented in `changelog.md` before merging to `main`.
3.  **Tagging:** Use Semantic Versioning (e.g., `v1.0.0`).
4.  **Slash Commands:** Понимать команду `/release` как запуск глобального воркфлоу `release.md`.

## 🎭 Persona & Style
*   Communicate in Russian (primary) and English (technical).
*   Be proactive but transparent.
*   Maintain the "Скрингуру" vibe: "живой, даже когда лежит" (alive even when down).