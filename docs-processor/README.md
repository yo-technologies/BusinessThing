# Document Processing — индексация и подготовка текстовых данных

## Назначение
Document Processing отвечает за асинхронную обработку загруженных документов: извлечение текста, разбиение на логические фрагменты и подготовку векторных представлений для последующего поиска релевантного контента.

## Технологический стек
- **Язык**: Go 1.24
- **Очереди**: RabbitMQ
- **Хранилище файлов**: S3
- **Векторная БД**: OpenSearch
- **Embeddings**: OpenAI API (text-embedding-3-small)
- **API**: gRPC + gRPC Gateway
- **Парсинг**: ledongthuc/pdf, custom TXT parser, docconv/v2
- **Логирование**: zap
- **Трейсинг**: Jaeger
- **Контейнеризация**: Docker + Docker Compose

## Структура проекта
```
docs-processor/
├── api/               # Protocol Buffers определения
│   └── document/
│       └── document.proto
├── cmd/               # Точки входа приложения
│   ├── doc-processor/ # gRPC сервер для поиска
│   │   └── main.go
│   └── worker/        # Воркер обработки документов
│       └── main.go
├── internal/
│   ├── app/           # Инициализация приложения
│   │   ├── api/       # gRPC handlers
│   │   └── app.go
│   ├── chunker/       # Разбиение текста на фрагменты
│   ├── config/        # Конфигурация
│   ├── domain/        # Доменные модели
│   ├── embeddings/    # Клиент для генерации embeddings
│   ├── logger/        # Логирование
│   ├── parser/        # Парсеры документов (PDF, TXT)
│   ├── queue/         # RabbitMQ клиент
│   ├── service/       # Бизнес-логика
│   ├── storage/       # S3 клиент
│   ├── tracer/        # Jaeger трейсинг
│   └── vectordb/      # OpenSearch клиент
├── pkg/               # Сгенерированный код из proto
├── config.yaml        # Конфигурация для локальной разработки
├── config.docker.yaml # Конфигурация для Docker
├── compose.yaml       # Docker Compose конфигурация
├── Dockerfile         # Образ Docker
└── Makefile          # Команды сборки и генерации
```

## Зона ответственности
- Получение задач на обработку документов из очереди RabbitMQ
- Чтение исходного файла из S3 хранилища
- Извлечение текстового содержимого из PDF / DOCX / TXT
- Разбиение текста на чанки с перекрытием для контекста
- Генерация векторных представлений (embeddings) через OpenAI API
- Индексация фрагментов в OpenSearch
- Предоставление gRPC API для поиска релевантных фрагментов

## Компоненты

### 1. Document Processor Worker
Асинхронный обработчик документов:
- Подписывается на очередь RabbitMQ
- Извлекает текст из документов
- Генерирует чанки с метаданными
- Создает embeddings батчами
- Индексирует в OpenSearch

### 2. gRPC API Server
Предоставляет API для поиска:
- `SearchChunks` - векторный поиск по документам организации
- HTTP Gateway на порту 8081
- Метрики Prometheus на `/metrics`

## Контракт событий обработки

Сервис читает задания из очереди RabbitMQ и ожидает сообщения в формате JSON.

- Очередь: `document_processing` (см. `rabbitmq.queue_name` в `config.yaml`)
- Content-Type: `application/json`
- Кодировка: UTF-8

### 1. Событие обработки документа (`job_type: "document"`)

Используется для индексации загруженных документов.

Обязательные поля:
- `job_type` — тип задачи, значение: `"document"`
- `document_id` — UUID документа
- `organization_id` — UUID организации
- `s3_key` — ключ файла в S3
- `document_type` — тип документа: `pdf` | `docx` | `txt`
- `document_name` — отображаемое имя файла
- `retry_count` — текущее количество попыток (обычно 0)
- `max_retries` — максимальное количество попыток (обычно 3)
- `created_at` — timestamp создания задачи в формате RFC3339

Пример сообщения:

```json
{
	"job_type": "document",
	"document_id": "123e4567-e89b-12d3-a456-426614174000",
	"organization_id": "456e7890-e89b-12d3-a456-426614174000",
	"s3_key": "documents/Даньшин Семён.pdf",
	"document_type": "pdf",
	"document_name": "Даньшин Семён.pdf",
	"retry_count": 0,
	"max_retries": 3,
	"created_at": "2025-11-17T12:00:00Z"
}
```

### 2. Событие индексации шаблона (`job_type: "template_index"`)

Используется для индексации метаданных шаблонов контрактов в векторную БД для семантического поиска.

Обязательные поля:
- `job_type` — тип задачи, значение: `"template_index"`
- `template_id` — UUID шаблона
- `template_name` — название шаблона
- `description` — описание шаблона
- `template_type` — тип шаблона (например, "contract", "agreement")
- `fields_count` — количество полей в шаблоне
- `retry_count` — текущее количество попыток (обычно 0)
- `max_retries` — максимальное количество попыток (обычно 3)
- `created_at` — timestamp создания задачи в формате RFC3339

Пример сообщения:

```json
{
	"job_type": "template_index",
	"template_id": "789e0123-e89b-12d3-a456-426614174000",
	"template_name": "Договор купли-продажи",
	"description": "Стандартный договор купли-продажи недвижимости",
	"template_type": "contract",
	"fields_count": 15,
	"retry_count": 0,
	"max_retries": 3,
	"created_at": "2025-11-17T12:00:00Z"
}
```

### 3. Событие удаления шаблона (`job_type: "template_delete"`)

Используется для удаления шаблона из векторного индекса.

Обязательные поля:
- `job_type` — тип задачи, значение: `"template_delete"`
- `template_id` — UUID шаблона для удаления
- `retry_count` — текущее количество попыток (обычно 0)
- `max_retries` — максимальное количество попыток (обычно 3)
- `created_at` — timestamp создания задачи в формате RFC3339

Пример сообщения:

```json
{
	"job_type": "template_delete",
	"template_id": "789e0123-e89b-12d3-a456-426614174000",
	"retry_count": 0,
	"max_retries": 3,
	"created_at": "2025-11-17T12:00:00Z"
}
```

### Примечания:
- Все UUID поля должны быть в формате RFC4122
- `s3_key` должен ссылаться на уже загруженный объект в указанном бакете (см. раздел `s3` в конфигурации)
- При достижении `max_retries` задача помечается как failed и не переобрабатывается
- Для больших файлов рекомендуется загружать в S3 заранее и публиковать событие только после успешной загрузки
- Поля с `null` значениями должны быть опущены (omitempty)
