# TCPDump to JMX Converter API 🚀

REST API сервис для конвертации TCP dump файлов в HAR и JMX форматы с автоматической корреляцией и параметризацией для Apache JMeter.

## 🌟 Возможности

- **Конвертация PCAP → HAR**: Преобразование TCP dump (PCAP) файлов в HTTP Archive формат
- **Конвертация HAR → JMX**: Генерация JMeter тест-планов из HAR файлов
- **Автоматическая корреляция**: Интеллектуальное обнаружение и корреляция динамических значений (session IDs, CSRF tokens, JWT и т.д.)
- **Параметризация**: Автоматическая параметризация тестовых данных
- **S3 хранилище**: Безопасное хранение конвертированных файлов в AWS S3 или MinIO
- **WebSocket поддержка**: Real-time обновления прогресса конвертации
- **REST API**: Простой и понятный API интерфейс
- **Docker поддержка**: Готовые Docker образы для быстрого развертывания

## 📋 Требования

- Go 1.21+
- Docker и Docker Compose (опционально)
- AWS S3 или MinIO для хранения файлов
- libpcap для парсинга PCAP файлов

## 🚀 Быстрый старт

### Использование Docker Compose (рекомендуется)

1. Клонируйте репозиторий:
```bash
git clone https://github.com/your-repo/tcpdump-to-jmx.git
cd tcpdump-to-jmx
```

2. Создайте `.env` файл:
```bash
cp .env.example .env
# Отредактируйте .env и укажите ваши AWS credentials
```

3. Запустите сервисы:
```bash
docker-compose up -d
```

Сервис будет доступен по адресу: http://localhost:8080

### Локальная установка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте `.env` файл:
```bash
cp .env.example .env
# Настройте переменные окружения
```

3. Запустите сервер:
```bash
go run main.go
```

## 📡 API Endpoints

### Конвертация файла

```bash
POST /api/v1/convert
```

Загрузка PCAP файла для конвертации:

```bash
curl -X POST http://localhost:8080/api/v1/convert \
  -F "file=@capture.pcap" \
  -F "correlation=true" \
  -F "parameterization=true" \
  -F "threads=10" \
  -F "rampup=10"
```

### Проверка статуса

```bash
GET /api/v1/status/{jobId}
```

### Скачивание результатов

```bash
GET /api/v1/download/{jobId}/har  # Скачать HAR файл
GET /api/v1/download/{jobId}/jmx  # Скачать JMX файл
```

### WebSocket для real-time обновлений

```javascript
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/${jobId}`);
ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    console.log(`Progress: ${update.progress}%`);
};
```

## 🔧 Конфигурация

### Переменные окружения

```env
# Server
SERVER_PORT=8080
GIN_MODE=release
MAX_FILE_SIZE=524288000  # 500MB

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
S3_BUCKET_NAME=tcpdump-conversions

# Для MinIO или S3-совместимых сервисов
S3_ENDPOINT=http://localhost:9000

# Workers
MAX_WORKERS=10
JOB_TIMEOUT=3600
RETENTION_PERIOD=168  # 7 дней
```

## 🎯 Автоматическая корреляция

Сервис автоматически обнаруживает и коррелирует:

- **Session IDs**: JSESSIONID, PHPSESSID, session_id
- **CSRF Tokens**: csrf_token, authenticity_token, xsrf_token
- **JWT Tokens**: Bearer tokens, JWT в headers и body
- **View States**: ASP.NET __VIEWSTATE
- **Custom IDs**: Любые динамические идентификаторы в URL и параметрах

## 📊 Параметризация

Автоматическая параметризация включает:

- Замена числовых ID в URL на переменные
- Параметризация UUID
- Замена credentials (username, password)
- Параметризация пагинации (page, offset, limit)
- Обработка динамических токенов

## 🐳 Docker

### Build образа

```bash
docker build -t tcpdump-to-jmx .
```

### Запуск контейнера

```bash
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  --name tcpdump-to-jmx \
  tcpdump-to-jmx
```

## 📚 Примеры использования

### Python клиент

```python
import requests
import websocket
import json

# Загрузка файла
with open('capture.pcap', 'rb') as f:
    response = requests.post(
        'http://localhost:8080/api/v1/convert',
        files={'file': f},
        data={
            'correlation': 'true',
            'parameterization': 'true'
        }
    )

job_data = response.json()
job_id = job_data['job_id']

# Подключение к WebSocket для отслеживания прогресса
def on_message(ws, message):
    update = json.loads(message)
    print(f"Progress: {update['progress']}% - {update['message']}")
    if update['status'] == 'completed':
        ws.close()

ws = websocket.WebSocketApp(
    f"ws://localhost:8080/api/v1/ws/{job_id}",
    on_message=on_message
)
ws.run_forever()

# Скачивание результатов
har_response = requests.get(f'http://localhost:8080/api/v1/download/{job_id}/har')
with open('output.har', 'wb') as f:
    f.write(har_response.content)

jmx_response = requests.get(f'http://localhost:8080/api/v1/download/{job_id}/jmx')
with open('output.jmx', 'wb') as f:
    f.write(jmx_response.content)
```

### Node.js клиент

```javascript
const FormData = require('form-data');
const fs = require('fs');
const axios = require('axios');
const WebSocket = require('ws');

// Загрузка файла
const form = new FormData();
form.append('file', fs.createReadStream('capture.pcap'));
form.append('correlation', 'true');
form.append('parameterization', 'true');

axios.post('http://localhost:8080/api/v1/convert', form, {
    headers: form.getHeaders()
}).then(response => {
    const jobId = response.data.job_id;
    
    // WebSocket для прогресса
    const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/${jobId}`);
    
    ws.on('message', (data) => {
        const update = JSON.parse(data);
        console.log(`Progress: ${update.progress}% - ${update.message}`);
        
        if (update.status === 'completed') {
            ws.close();
            downloadResults(jobId);
        }
    });
});

function downloadResults(jobId) {
    // Скачивание HAR
    axios.get(`http://localhost:8080/api/v1/download/${jobId}/har`, {
        responseType: 'stream'
    }).then(response => {
        response.data.pipe(fs.createWriteStream('output.har'));
    });
    
    // Скачивание JMX
    axios.get(`http://localhost:8080/api/v1/download/${jobId}/jmx`, {
        responseType: 'stream'
    }).then(response => {
        response.data.pipe(fs.createWriteStream('output.jmx'));
    });
}
```

## 🏗️ Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   REST API  │────▶│   Worker    │
└─────────────┘     └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  WebSocket  │     │ Converters  │
                    └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │   Progress  │     │   AWS S3    │
                    └─────────────┘     └─────────────┘
```

## 📦 Структура проекта

```
tcpdump-to-jmx/
├── main.go                 # Entry point
├── internal/
│   ├── api/               # REST API handlers
│   ├── config/            # Configuration
│   ├── converter/         # PCAP→HAR, HAR→JMX converters
│   ├── models/            # Data models
│   ├── storage/           # S3 storage interface
│   └── worker/            # Job processing
├── docs/                  # API documentation
├── Dockerfile            # Docker image
├── docker-compose.yml    # Docker Compose config
├── go.mod               # Go dependencies
└── README.md           # This file
```

## 🤝 Contributing

Приветствуются pull requests. Для больших изменений, пожалуйста, откройте issue для обсуждения.

## 📄 License

MIT License - см. файл LICENSE для деталей.

## 🙏 Acknowledgments

- [gopacket](https://github.com/google/gopacket) - для парсинга PCAP
- [Gin](https://github.com/gin-gonic/gin) - веб-фреймворк
- [AWS SDK](https://github.com/aws/aws-sdk-go) - для работы с S3
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket поддержка