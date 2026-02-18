---
description: Оптимизированный процесс релиза для Скрингуру.
---
// turbo-all

Этот воркфлоу оптимизирован для быстрой публикации версий.

### 1. Подготовка и Коммит
1. **Версия:** Определить новую версию (X.Y.Z) на основе `changelog.md`.
2. **Фиксация:** Убедиться, что `changelog.md` обновлен и все изменения закоммичены в `dev`.
   ```bash
   git add .
   git commit -m "chore: prepare release {vX.X.X}"
   git push origin dev
   ```

### 2. Слияние в Main и Тегирование
1. **Merge:** Переключиться на `main`, подтянуть изменения и влить `dev`.
   ```bash
   git checkout main
   git pull origin main
   git merge dev --no-ff -m "Release {vX.X.X}"
   git push origin main
2. **Tag:** Создать и отправить тег.
   ```bash
   git tag {vX.X.X}
   git push origin {vX.X.X}
   ```

### 3. Публикация на GitHub
1. **Release Notes:** Извлечь последнюю запись из ченджлога и создать релиз.
   ```powershell
   $notes = (Get-Content changelog.md -Raw -Encoding UTF8 | Select-String -Pattern '(?s)##\s+\[.*?\].*?(?=##\s+\[|$)').Matches.Value; [System.IO.File]::WriteAllText("RELEASENOTES.tmp", $notes, (New-Object System.Text.UTF8Encoding $false)); gh release create {vX.X.X} --title "{vX.X.X}" --notes-file RELEASENOTES.tmp; Remove-Item RELEASENOTES.tmp
   ```

### 4. Синхронизация Dev (ВАЖНО)
1. **Update Remote Dev:** После вливания в `main` (где создается merge-коммит), нужно обновить локальную и удаленную ветку `dev`, чтобы они не отставали.
   ```bash
   git checkout dev
   git merge main
   git push origin dev
   ```

---
**Примечание:** Теперь удаленный `dev` всегда будет соответствовать `main` после завершения релиза.
