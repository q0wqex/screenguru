---
description: Универсальный процесс релиза для любого Git-проекта.
---
// turbo-all

Этот воркфлоу — золотой стандарт для выпуска стабильной версии. Он подходит для любого проекта, использующего Git и GitHub.

### 1. Подготовка и Анализ
1. **Проверка окружения:** Убедиться, что нет незакоммиченных данных: `git status`.
2. **Ветка разработки:** Переключиться на ветку, где велась работа: `git checkout dev`.
3. **История:** Просмотреть список коммитов с момента последней версии: `git log $(git describe --tags --abbrev=0)..HEAD --oneline`.
4. **Определение версии:** Выбрать версию согласно SemVer (X.Y.Z).

### 2. Документирование (Changelog)
1. **Обновление:** Добавить новую запись в `changelog.md`.
2. **Фиксация документации:** 
   ```bash
   git add changelog.md
   git commit -m "docs: release version {vX.X.X}"
   ```

### 3. Слияние и Синхронизация
1. **Push:** Отправить изменения в удаленный репозиторий: `git push origin dev`.
2. **Merge в Stable:** 
   - Перейти в основную ветку: `git checkout main`.
   - Обновить её: `git pull origin main`.
   - Влить изменения: `git merge dev --no-ff -m "Merge branch 'dev' into main for release {vX.X.X}"`.
3. **Push Stable:** `git push origin main`.

### 4. Тегирование и GitHub Release
1. **Тег:** Создать аннотированный тег: `git tag -a {vX.X.X} -m "{vX.X.X}"`.
2. **Push тега:** `git push origin {vX.X.X}`.
3. **Подготовка лога релиза:** Извлечь только последнюю запись из `changelog.md` с корректной кодировкой UTF-8:
   ```powershell
   $utf8NoBom = New-Object System.Text.UTF8Encoding $false; $content = Get-Content changelog.md -Raw -Encoding UTF8; $notes = [regex]::Match($content, '(?s)##\s+\[.*?\].*?(?=##\s+\[|$)').Value; [System.IO.File]::WriteAllText("RELEASENOTES.tmp", $notes, $utf8NoBom)
   ```
4. **Release:** Создать официальный релиз на GitHub:
   ```bash
   gh release create {vX.X.X} --title "{vX.X.X}" --notes-file RELEASENOTES.tmp
   ```
5. **Очистка:** Удалить временный файл:
   ```powershell
   Remove-Item RELEASENOTES.tmp
   ```

### 5. Возврат к разработке
1. Вернуться в рабочую ветку: `git checkout dev`.
2. Влить изменения из main: `git merge main`.

---
**Примечание:** Этот воркфлоу автоматически извлекает актуальную часть ченджлога в правильной кодировке.
