---
description: Оптимизированный процесс релиза для Скрингуру.
---
// turbo-all

Этот воркфлоу оптимизирован для быстрой публикации версий.

### 1. Подготовка и Коммит
1. **Версия:** Определить новую версию (X.Y.Z) на основе `changelog.md` (или инкрементально).
2. **Фиксация:** Убедиться, что `changelog.md` обновлен (только пользовательские изменения!) и все изменения закоммичены в `dev`.
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
1. **Release Notes:** Сформировать описание (из `RELEASES.md` + Git Log) и создать релиз.
   ```powershell
   $notes = (Get-Content RELEASES.md -Raw -Encoding UTF8 | Select-String -Pattern '(?s)##\s+\[.*?\].*?(?=##\s+\[|$)').Matches[0].Value;
   $prevTag = git describe --tags --abbrev=0 HEAD^ 2>$null;
   $range = if ($prevTag) { "$prevTag..HEAD" } else { "HEAD" };
   $commits = git log $range --pretty=format:"* %h %s";
   $fullNotes = "$notes`n`n### 🛠 Commits`n$commits";
   [System.IO.File]::WriteAllText("RELEASENOTES.tmp", $fullNotes, (New-Object System.Text.UTF8Encoding $false));
   gh release create {vX.X.X} --title "{vX.X.X}" --notes-file RELEASENOTES.tmp;
   Remove-Item RELEASENOTES.tmp
   ```

### 4. Синхронизация Dev (ВАЖНО)
1. **Update Remote Dev:** После вливания в `main` (где создается merge-коммит), нужно обновить локальную и удаленную ветку `dev`, чтобы они не отставали.
   ```bash
   git checkout dev
   git merge main
   git push origin dev
   ```

---
**Примечание:**
- `changelog.md`: Только для пользователей (красиво, кратко). Тянется в UI сайта.
- `RELEASES.md`: Технический лог. Используется для формирования Release Notes на GitHub (+ авто-список коммитов).
