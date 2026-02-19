---
description: Оптимизированный процесс релиза для Скрингуру (Smart & Ephemeral).
---

// turbo-all

Этот воркфлоу полагается на аналитические способности Antigravity для создания релизов без мусора.

### 1. Подготовка и Анализ (Interactive)
1. **Discovery:** Antigravity находит последний тег и анализирует коммиты с момента последней публикации.
   - `git describe --tags --abbrev=0`
   - `git log <last_tag>..HEAD --oneline`
2. **Drafting:** Antigravity предлагает:
   - Новую версию (Semantic Versioning).
   - Текст для `changelog.md` (на русском, ориентирован на пользователя) — **опционально**, пользователь может отказаться.
   - Технический текст для GitHub Release (на русском и английском).
3. **Approval:** Пользователь подтверждает или вносит правки через `notify_user`.

### 2. Подготовка файлов
1. **Changelog:** Дописать новый блок изменений в начало `changelog.md` — **только если пользователь подтвердил**. Если пользователь отказался («не надо», «пропусти»), шаг пропускается.
2. **Commit:**
   ```powershell
   git add .
   git commit -m "chore: prepare release {vX.X.X}"
   git push origin dev
   ```

### 3. Слияние и Тегирование
1. **Merge (squash):** Все коммиты из `dev` схлопываются в один чистый коммит на `main`.
   ```powershell
   git checkout main
   git pull origin main
   git merge dev --squash
   git commit -m "Release {vX.X.X}"
   git push origin main
   ```
2. **Tag:**
   ```powershell
   git tag {vX.X.X}
   git push origin {vX.X.X}
   ```

### 4. Публикация на GitHub
1. **Notes:** Создать временный файл `RELEASENOTES.tmp` с техническим описанием.
   - ⚠️ **ВАЖНО:** Файл создавать **только через `write_to_file` tool**, НЕ через `Set-Content` в PowerShell — иначе кодировка будет сломана и первые буквы кириллических строк потеряются.
2. **Create:**
   ```powershell
   # Обязательно устанавливаем UTF-8 перед вызовом gh!
   [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
   gh release create {vX.X.X} --title "{vX.X.X}" --notes-file RELEASENOTES.tmp
   ```
3. **Verify:** Проверить что notes залились корректно (первые буквы на месте):
   ```powershell
   [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
   gh release view {vX.X.X} --json body -q .body
   ```
   Если буквы пропали — исправить через:
   ```powershell
   [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
   gh release edit {vX.X.X} --notes-file RELEASENOTES.tmp
   ```
4. **Cleanup:** Удалить `RELEASENOTES.tmp`.

### 5. Синхронизация
1. **Sync Dev:** После squash-merge `dev` и `main` расходятся по истории — нужно явно смержить `main` обратно в `dev`.
   ```powershell
   git checkout dev
   git merge main --no-ff -m "chore: sync dev with main after release {vX.X.X}"
   git push origin dev
   ```

---
**Примечание:**
- **changelog.md**: Теперь хранит всю историю изменений для пользователей.
- **Git History на `main`**: Только чистые `Release vX.X.X` коммиты — один на релиз.
- **Git History на `dev`**: Полная история со всеми `chore:` коммитами.
- **Squash merge**: Все коммиты разработки "схлопываются" в один коммит при слиянии в `main`.