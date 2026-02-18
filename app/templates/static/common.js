document.addEventListener('DOMContentLoaded', function () {
  const uploadArea = document.getElementById('uploadArea');
  const fileInput = document.getElementById('fileInput');
  const uploadForm = document.getElementById('uploadForm') || document.getElementById('imageUploadForm');

  if (uploadArea && fileInput && uploadForm) {
    uploadArea.addEventListener('click', function () { fileInput.click() }, { passive: true });
    fileInput.addEventListener('change', function () {
      if (fileInput.files.length > 0) {
        handleUpload(fileInput.files, uploadForm);
      }
    });
    uploadArea.addEventListener('dragover', function (e) {
      e.preventDefault();
      uploadArea.classList.add('dragover');
    });
    uploadArea.addEventListener('dragleave', function (e) {
      e.preventDefault();
      uploadArea.classList.remove('dragover');
    });
    uploadArea.addEventListener('drop', function (e) {
      e.preventDefault();
      uploadArea.classList.remove('dragover');
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        fileInput.files = files;
        handleUpload(files, uploadForm);
      }
    });
  }

  // Инициализация темы
  initTheme();

});

// Функции для работы с темами
function initTheme() {
  const savedTheme = localStorage.getItem('screenguru_theme') || 'crystal';
  applyTheme(savedTheme);

  const themeSelect = document.getElementById('themeSelect');
  if (themeSelect) {
    themeSelect.value = savedTheme;
  }
}

function changeTheme(themeName) {
  applyTheme(themeName);
  localStorage.setItem('screenguru_theme', themeName);
}

function applyTheme(themeName) {
  if (themeName === 'crystal') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', themeName);
  }
  // Иконки теперь вставлены инлайново или заменены на эмодзи, перерисовка не нужна
}

// Показать индикатор загрузки
function showUploadProgress(total) {
  const overlay = document.getElementById('uploadOverlay');
  const status = document.getElementById('uploadStatus');
  const count = document.getElementById('uploadCount');

  overlay.classList.add('active');
  status.textContent = 'зᴀᴦᴩузᴋᴀ...';
  count.textContent = '0 / ' + total + ' файлов';

  return {
    update: function (current) {
      status.textContent = 'зᴀᴦᴩузᴋᴀ...';
      count.textContent = current + ' / ' + total + ' файлов';
    },
    hide: function () {
      overlay.classList.remove('active');
    }
  };
}

// handleUpload обрабатывает загрузку файлов
function handleUpload(files, form) {
  const albumInput = form.querySelector('input[name="album_id"]');

  // Если album_id уже есть в форме (загрузка в существующий альбом)
  if (albumInput && albumInput.value) {
    // sessionID из URL текущей страницы
    const pathParts = window.location.pathname.split('/').filter(p => p);
    const sessionID = pathParts[0] || '';
    uploadFilesParallel(files, albumInput.value, sessionID);
    return;
  }

  // Иначе создаем новый альбом на сервере
  fetch('/create-album', {
    method: 'POST',
    credentials: 'same-origin'
  })
    .then(response => response.json())
    .then(data => {
      if (data.album_id && data.session_id) {
        uploadFilesParallel(files, data.album_id, data.session_id);
      } else {
        throw new Error('Failed to create album');
      }
    })
    .catch(error => {
      console.error('Error creating album:', error);
      alert('Ошибка при создании альбома');
    });
}

// uploadFilesParallel отправляет файлы параллельно небольшими пачками
function uploadFilesParallel(files, albumID, sessionID) {
  const total = files.length;
  let completed = 0;
  const progress = showUploadProgress(total);
  const concurrencyLimit = 3; // Загружаем по 3 файла одновременно
  let currentIndex = 0;

  async function uploadWorker() {
    while (currentIndex < total) {
      const index = currentIndex++;
      const file = files[index];

      try {
        // Конвертируем в WebP перед загрузкой
        let uploadFile = file;
        try {
          uploadFile = await convertToWebP(file);
        } catch (convErr) {
          console.warn('Conversion failed, uploading original:', convErr);
        }

        const formData = new FormData();
        formData.append('image', uploadFile);
        formData.append('album_id', albumID);

        const response = await fetch('/upload', {
          method: 'POST',
          body: formData,
          credentials: 'same-origin',
          headers: {
            'X-Requested-With': 'XMLHttpRequest'
          }
        });

        if (!response.ok) {
          throw new Error('Upload failed for ' + file.name);
        }

        completed++;
        progress.update(completed);
      } catch (error) {
        console.error('Upload error:', error);
        throw error;
      }
    }
  }

  // Запускаем воркеры
  const workers = [];
  const numWorkers = Math.min(concurrencyLimit, total);
  for (let i = 0; i < numWorkers; i++) {
    workers.push(uploadWorker());
  }

  Promise.all(workers)
    .then(() => {
      progress.hide();
      window.location.href = '/' + sessionID + '/' + albumID;
    })
    .catch(error => {
      progress.hide();
      console.error('Batch upload error:', error);
      alert('Ошибка при загрузке: ' + error.message);
    });
}

// getSessionID получает ID сессии из cookie
function getSessionID() {
  const cookies = document.cookie.split(';');
  for (let i = 0; i < cookies.length; i++) {
    const cookie = cookies[i].trim();
    if (cookie.indexOf('session_id=') === 0) {
      return cookie.substring('session_id='.length, cookie.length);
    }
  }
  return '';
}

const EMPTY_STATE_HTML = `
  <div class="empty-state">
    <div class="empty-icon">📂</div>
    <div class="empty-text">у вас пока нет загруженных изображений</div>
  </div>
`;

// showCopiedFeedback показывает визуальную обратную связь о копировании
function showCopiedFeedback(button) {
  const originalText = button.textContent;
  button.textContent = 'ᴄᴋоᴨиᴩоʙᴀно!';
  button.classList.add('copied');
  setTimeout(function () {
    button.textContent = originalText;
    button.classList.remove('copied');
  }, 2000);
}

// Функция для копирования ссылки на альбом
function copyAlbumUrl(sessionID, albumID, button) {
  const url = window.location.origin + '/' + sessionID + '/' + albumID;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url)
      .then(function () { showCopiedFeedback(button) })
      .catch(function (err) { console.error('нᴇ удᴀᴧоᴄь ᴄᴋоᴨиᴩоʙᴀᴛь ᴜʀʟ: ', err) });
  } else {
    // Fallback для старых браузеров
    const textArea = document.createElement('textarea');
    textArea.value = url;
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      document.execCommand('copy');
      showCopiedFeedback(button);
    } catch (err) { console.error('Не удалось скопировать URL: ', err) }
    document.body.removeChild(textArea);
  }
}

function copyUrl(sessionID, albumID, filename, button) {
  const url = window.location.origin + '/' + sessionID + '/' + albumID + '/' + filename;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url)
      .then(function () { showCopiedFeedback(button) })
      .catch(function (err) { console.error('нᴇ удᴀᴧоᴄь ᴄᴋоᴨиᴩоʙᴀᴛь ᴜʀʟ: ', err) });
  } else {
    // Fallback для старых браузеров
    const textArea = document.createElement('textarea');
    textArea.value = url;
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      document.execCommand('copy');
      showCopiedFeedback(button);
    } catch (err) { console.error('Не удалось скопировать URL: ', err) }
    document.body.removeChild(textArea);
  }
}

function deleteImage(sessionID, albumID, filename, button) {
  if (!confirm('Вы уверены, что хотите удалить это изображение?')) {
    return;
  }

  const formData = new FormData();
  formData.append('album_id', albumID);
  formData.append('filename', filename);

  fetch('/delete-image', {
    method: 'POST',
    body: formData
  })
    .then(response => {
      if (response.ok) {
        // Удаляем элемент изображения из DOM
        const imageItem = button.closest('.image-item');
        imageItem.style.transition = 'opacity 0.3s ease';
        imageItem.style.opacity = '0';
        setTimeout(() => {
          imageItem.remove();
          // Проверяем, остались ли изображения
          const imageGrid = document.getElementById('imageGrid');
          if (!imageGrid.querySelector('.image-item')) {
            // Показываем пустое состояние
            imageGrid.innerHTML = EMPTY_STATE_HTML;
          }
        }, 300);
      } else {
        alert('Ошибка при удалении изображения');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      alert('Ошибка при удалении изображения');
    });
}

function deleteUser() {
  if (!confirm('Вы уверены, что хотите удалить весь профиль со всеми альбомами и изображениями? Это действие необратимо!')) {
    return;
  }

  fetch('/delete-user', {
    method: 'POST'
  })
    .then(response => {
      if (response.ok) {
        // Перезагружаем страницу - сервер уже очистил cookie
        window.location.href = '/';
      } else {
        alert('Ошибка при удалении профиля');
      }
    })
    .catch(error => {
      console.error('Error:', error);
      alert('Ошибка при удалении профиля');
    });
}

// Открывает изображение в оверлее
function toggleZoom(img) {
  const overlay = document.getElementById('image-viewer-overlay');
  const zoomedImageContainer = document.getElementById('zoomed-image-element');

  // Если оверлей уже активен с этим же изображением, ничего не делаем
  if (overlay.classList.contains('active') && zoomedImageContainer.firstChild && zoomedImageContainer.firstChild.src === img.src) {
    return;
  }

  // Используем Image объект для предзагрузки перед показом
  const newImg = new Image();
  newImg.src = img.src;
  newImg.alt = img.alt;

  newImg.onload = function () {
    zoomedImageContainer.innerHTML = '';
    zoomedImageContainer.appendChild(newImg);
    overlay.classList.add('active');
  };
}

// Закрывает оверлей
function closeZoom() {
  const overlay = document.getElementById('image-viewer-overlay');
  overlay.classList.remove('active');
}


// convertToWebP конвертирует изображение в формат WebP
function convertToWebP(file) {
  return new Promise((resolve, reject) => {
    // Проверяем, является ли файл изображением
    if (!file.type.startsWith('image/')) {
      // Если это не изображение, просто возвращаем исходный файл
      resolve(file);
      return;
    }

    // Создаем объект FileReader для чтения файла
    const reader = new FileReader();
    reader.onload = function (e) {
      // Создаем элемент img для загрузки изображения
      const img = new Image();
      img.onload = function () {
        // Создаем canvas элемент для конвертации
        const canvas = document.createElement('canvas');
        canvas.width = img.width;
        canvas.height = img.height;

        const ctx = canvas.getContext('2d');
        // Рисуем изображение на canvas
        ctx.drawImage(img, 0, 0);

        // Конвертируем canvas в WebP формат
        canvas.toBlob(function (blob) {
          if (blob) {
            // Создаем новый File объект с правильным именем и типом
            const fileName = file.name.replace(/\.[^/.]+$/, '') + '.webp';
            const webpFile = new File([blob], fileName, { type: 'image/webp' });
            resolve(webpFile);
          } else {
            reject(new Error('Failed to convert image to WebP'));
          }
        }, 'image/webp', 0.82); // Качество 82% (чуть снизили для скорости)
      };
      img.onerror = function () {
        reject(new Error('Failed to load image'));
      };
      img.src = e.target.result;
    };
    reader.onerror = function () {
      reject(new Error('Failed to read file'));
    };
    reader.readAsDataURL(file);
  });
}

// Сравнение семантических версий (v1 > v2 => 1, v1 < v2 => -1, v1 == v2 => 0)
function compareVersions(v1, v2) {
  if (!v1) return 1;
  if (!v2) return 1;
  const a = v1.split('.').map(Number);
  const b = v2.split('.').map(Number);
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    const na = a[i] || 0;
    const nb = b[i] || 0;
    if (na > nb) return 1;
    if (na < nb) return -1;
  }
  return 0;
}

// Работа с ченджлогом
function checkChangelog() {
  fetch('/changelog')
    .then(response => response.json())
    .then(data => {
      if (!data.data || !data.data.content) return;

      const content = data.data.content;
      // Находим все версии (заголовки ## [X.X.X])
      const versionMatch = content.match(/## \[?([\d.]+)\]?/);
      if (!versionMatch) return;

      const latestVersion = versionMatch[1];
      const savedVersion = localStorage.getItem('last_seen_version');

      // Если последняя версия новее сохраненной
      if (compareVersions(latestVersion, savedVersion) > 0) {
        showChangelog(content, latestVersion, savedVersion);
      }
    })
    .catch(error => console.error('Error fetching changelog:', error));
}

function showChangelog(content, latestVersion, savedVersion) {
  const modal = document.getElementById('changelogModal');
  const body = document.getElementById('changelogBody');

  if (!modal || !body) return;

  // Регулярка для поиска заголовков версий
  const versionRegex = /## \[?([\d.]+)\]?[^\n]*/g;
  const matches = [];
  let match;

  while ((match = versionRegex.exec(content)) !== null) {
    matches.push({
      version: match[1],
      header: match[0],
      index: match.index
    });
  }

  let fullHtml = '';
  let processedVersions = 0;

  for (let i = 0; i < matches.length; i++) {
    const currentMatch = matches[i];

    // Показываем только если версия новее сохраненной
    if (compareVersions(currentMatch.version, savedVersion) > 0) {
      const nextIndex = (i + 1 < matches.length) ? matches[i + 1].index : content.length;
      let sectionContent = content.substring(currentMatch.index + currentMatch.header.length, nextIndex).trim();

      // Очистка от разделителей
      sectionContent = sectionContent.replace(/---/g, '');

      let html = sectionContent
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>') // Жирный текст
        .replace(/`(.*?)`/g, '<code>$1</code>')         // Инлайн код
        .replace(/^### (.*$)/gim, '<h3>$1</h3>')
        .replace(/^\- (.*$)/gim, '<li>$1</li>');

      // Группируем li в ul (только внутри текущей секции)
      html = html.replace(/(<li>.*<\/li>(\s*<li>.*<\/li>)*)/g, '<ul>$1</ul>');

      fullHtml += `
        <div class="changelog-version-section">
          <div class="changelog-version-tag">Версия ${currentMatch.version}</div>
          ${html}
        </div>
      `;
      processedVersions++;
    } else {
      // Версии обычно идут по убыванию, можно остановиться
      break;
    }
  }

  if (processedVersions === 0) return;

  body.innerHTML = fullHtml;
  modal.dataset.version = latestVersion;
  modal.classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeChangelog() {
  const modal = document.getElementById('changelogModal');
  const version = modal.dataset.version;

  if (version) {
    localStorage.setItem('last_seen_version', version);
  }

  modal.classList.remove('active');
  document.body.style.overflow = '';
}

// Инициализация при загрузке
document.addEventListener('DOMContentLoaded', function () {
  // Вызываем проверку ченджлога через небольшую задержку для плавности
  setTimeout(checkChangelog, 1000);

  // Инициализация секретного HUD
  initSecretHUD();
});

// Секретный интерактив на F8
function initSecretHUD() {
  // 1. Создание элементов
  const flash = document.createElement('div');
  flash.className = 'screenshot-flash';

  const chat = document.createElement('div');
  chat.className = 'pseudo-chat';
  chat.innerHTML = `<span class="chat-prefix">></span><span id="chat-input-text"></span><span class="typing-cursor"></span>`;

  document.body.appendChild(flash);
  document.body.appendChild(chat);

  let isRunning = false;

  document.addEventListener('keydown', function (e) {
    if (e.key === 'F8' && !isRunning) {
      e.preventDefault();
      runSecretFlow();
    }
  });

  async function runSecretFlow() {
    isRunning = true;

    // 1. Показываем чат
    chat.style.display = 'block';
    const inputText = document.getElementById('chat-input-text');
    inputText.textContent = '';

    const command = '/время';
    for (let i = 0; i < command.length; i++) {
      await new Promise(r => setTimeout(r, 100)); // Тайпинг
      inputText.textContent += command[i];
    }

    await new Promise(r => setTimeout(r, 400));

    // 2. Вывод времени МСК
    const mskTime = new Date(new Date().getTime() + (new Date().getTimezoneOffset() * 60000) + (3 * 3600000));
    const timeStr = mskTime.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

    inputText.innerHTML = `<span id="msk-time-display">${timeStr} (MSK)</span>`;

    await new Promise(r => setTimeout(r, 600));

    // 3. Вспышка (скриншот)
    flash.classList.add('active');
    setTimeout(() => flash.classList.remove('active'), 500);

    await new Promise(r => setTimeout(r, 600));

    // 4. Выбор файлов
    const fileInput = document.getElementById('fileInput');
    if (fileInput) {
      fileInput.click();
    }

    // Очистка
    setTimeout(() => {
      chat.style.display = 'none';
      isRunning = false;
    }, 2000);
  }
}
