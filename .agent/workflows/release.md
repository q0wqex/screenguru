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
   - Текст для `changelog.md` (на русском, ориентирован на пользователя).
   - Технический текст для GitHub Release (на русском и английском).
3. **Approval:** Пользователь подтверждает или вносит правки через `notify_user`.

### 2. Подготовка файлов
1. **Changelog:** Дописать новый блок изменений в начало `changelog.md`.
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
2. **Create:**
   ```powershell
   gh release create {vX.X.X} --title "{vX.X.X}" --notes-file RELEASENOTES.tmp
   ```
3. **Cleanup:** Удалить `RELEASENOTES.tmp`.

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