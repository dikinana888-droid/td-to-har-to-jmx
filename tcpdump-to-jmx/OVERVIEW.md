# 🚀 TCPDump to JMX Converter - Обзор проекта

## 📁 Структура проекта

```
tcpdump-to-jmx/
├── main.go                     # Точка входа приложения
├── go.mod                      # Go модули и зависимости
├── .env.example                # Пример конфигурации
├── Dockerfile                  # Docker образ
├── docker-compose.yml          # Docker Compose конфигурация
├── Makefile                    # Команды для разработки
├── run.sh                      # Скрипт быстрого запуска
├── README.md                   # Документация
│
├── internal/                   # Внутренние пакеты
│   ├── api/                   # REST API обработчики
│   │   ├── handler.go         # Основные API эндпоинты
│   │   └── middleware.go      # Middleware (CORS, логирование)
│   │
│   ├── config/                # Конфигурация
│   │   └── config.go          # Загрузка конфигурации из ENV
│   │
│   ├── converter/             # Конвертеры
│   │   ├── pcap_to_har.go    # PCAP → HAR конвертер
│   │   ├── har_to_jmx.go     # HAR → JMX конвертер
│   │   └── jmx_models.go     # JMX структуры данных
│   │
│   ├── models/                # Модели данных
│   │   ├── job.go            # Модель задания
│   │   └── har.go            # HAR формат структуры
│   │
│   ├── storage/               # Хранилище
│   │   └── s3.go             # AWS S3 интеграция
│   │
│   └── worker/                # Обработка заданий
│       └── job_manager.go     # Менеджер заданий
│
├── docs/                       # Документация
│   └── index.html             # API документация (UI)
│
└── examples/                   # Примеры использования
    ├── python_client.py       # Python клиент
    ├── nodejs_client.js       # Node.js клиент
    └── package.json           # NPM зависимости

```

## 🎯 Основные возможности

### 1. **Конвертация PCAP → HAR**
- Парсинг TCP dump файлов (PCAP, PCAPNG, CAP)
- Извлечение HTTP/HTTPS трафика
- Реконструкция TCP потоков
- Формирование HAR структуры

### 2. **Конвертация HAR → JMX**
- Генерация JMeter тест-планов
- Создание Thread Groups
- HTTP Samplers для каждого запроса
- Добавление Cookie и Header менеджеров

### 3. **Автоматическая корреляция**
Система автоматически обнаруживает и коррелирует:
- **Session IDs**: JSESSIONID, PHPSESSID, session_id
- **CSRF токены**: csrf_token, authenticity_token, xsrf_token
- **JWT токены**: Bearer tokens в заголовках и теле
- **ViewState**: ASP.NET __VIEWSTATE
- **Динамические ID**: в URL путях и параметрах

### 4. **Параметризация**
Автоматическая замена на переменные:
- Числовые ID → `${id}`
- UUID → `${uuid}`
- Credentials → `${username}`, `${password}`
- Пагинация → `${page}`, `${offset}`, `${limit}`
- Токены → `${token_name}`

### 5. **S3 хранилище**
- Сохранение оригинальных PCAP файлов
- Хранение сгенерированных HAR файлов
- Хранение JMX файлов
- Поддержка AWS S3 и MinIO
- Автоматическая очистка старых файлов

### 6. **WebSocket поддержка**
- Real-time обновления прогресса
- Статус конвертации
- Сообщения об ошибках
- Процент выполнения

## 🔧 Технологический стек

- **Go 1.21**: Основной язык разработки
- **Gin**: Web фреймворк для REST API
- **gopacket**: Парсинг PCAP файлов
- **AWS SDK**: Интеграция с S3
- **Gorilla WebSocket**: WebSocket поддержка
- **Docker**: Контейнеризация
- **MinIO**: Локальное S3-совместимое хранилище

## 🚀 Быстрый старт

### 1. Локальный запуск

```bash
# Установка зависимостей
go mod download

# Сборка
./run.sh build

# Запуск
./run.sh run
```

### 2. Docker Compose (рекомендуется)

```bash
# Запуск всех сервисов
./run.sh compose-up

# API будет доступен на http://localhost:8080
# MinIO консоль на http://localhost:9001
```

### 3. Использование API

```bash
# Проверка здоровья
curl http://localhost:8080/api/v1/health

# Конвертация файла
curl -X POST http://localhost:8080/api/v1/convert \
  -F "file=@capture.pcap" \
  -F "correlation=true" \
  -F "parameterization=true"

# Проверка статуса
curl http://localhost:8080/api/v1/status/{job_id}

# Скачивание результатов
curl -O http://localhost:8080/api/v1/download/{job_id}/har
curl -O http://localhost:8080/api/v1/download/{job_id}/jmx
```

## 📊 API Endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/health` | Проверка состояния API |
| POST | `/api/v1/convert` | Загрузка и конвертация PCAP файла |
| GET | `/api/v1/status/:jobId` | Статус задания |
| GET | `/api/v1/download/:jobId/:type` | Скачивание файла (har/jmx) |
| GET | `/api/v1/ws/:jobId` | WebSocket для прогресса |
| GET | `/api/v1/conversions` | Список конвертаций |

## 🔐 Конфигурация

Основные переменные окружения:

```env
# Сервер
SERVER_PORT=8080
MAX_FILE_SIZE=524288000  # 500MB

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
S3_BUCKET_NAME=tcpdump-conversions

# Для локальной разработки (MinIO)
S3_ENDPOINT=http://localhost:9000

# Воркеры
MAX_WORKERS=10
JOB_TIMEOUT=3600
RETENTION_PERIOD=168  # 7 дней
```

## 📝 Примеры использования

### Python
```python
from examples.python_client import TCPDumpToJMXClient

client = TCPDumpToJMXClient("http://localhost:8080")
job_id = client.convert_file("capture.pcap")
status = client.wait_for_completion(job_id)
client.download_file(job_id, "jmx")
```

### Node.js
```javascript
const TCPDumpToJMXClient = require('./examples/nodejs_client');

const client = new TCPDumpToJMXClient('http://localhost:8080');
const jobId = await client.convertFile('capture.pcap');
await client.monitorProgress(jobId);
await client.downloadFile(jobId, 'jmx');
```

## 🧪 Тестирование

```bash
# Запуск тестов
go test ./... -v

# Тестовый скрипт
./test_api.sh
```

## 🐳 Docker

```bash
# Сборка образа
docker build -t tcpdump-to-jmx .

# Запуск контейнера
docker run -p 8080:8080 --env-file .env tcpdump-to-jmx

# Docker Compose
docker-compose up -d
```

## 📚 Дополнительные возможности

1. **Фильтрация трафика**
   - По порту: `port=8080`
   - По хосту: `host=example.com`

2. **Настройка JMeter**
   - Количество потоков: `threads=10`
   - Время разгона: `rampup=10`
   - Количество итераций: `loops=1`

3. **Управление корреляцией**
   - Включить: `correlation=true`
   - Выключить: `correlation=false`

4. **Управление параметризацией**
   - Включить: `parameterization=true`
   - Выключить: `parameterization=false`

## 🤝 Поддержка

- Документация API: http://localhost:8080/docs
- GitHub Issues: [Создать issue](https://github.com/your-repo/issues)
- Email: support@example.com

## 📄 Лицензия

MIT License - свободное использование в коммерческих и некоммерческих проектах.

---

**Версия**: 1.0.0  
**Автор**: AI Assistant  
**Дата создания**: 2024