# Loyalty-service

Накопительная система лояльности

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.

### Нейминг веток

- **feat/** — новая функциональность
- **fix/** — исправление ошибки
- **docs/** — изменения в документации
- **style/** — форматирование кода, отступы, точки с запятой (без изменения логики)
- **refactor/** — рефакторинг кода (не исправляет ошибку и не добавляет функциональность)
- **perf/** — изменение, улучшающее производительность
- **test/** — добавление или обновление тестов
- **chore/** — рутинные задачи: обновление зависимостей, настройка сборки
- **ci/** — изменения в CI/CD конфигурации
- **build/** — изменения в системе сборки или внешних зависимостях

### Примеры
feat/#2-registration
fix/#15-login-validation
docs/#7-api-endpoints
chore/#3-docker-setup
test/#12-auth-handlers
refactor/#18-database-layer

### Примечания

- Все ветки вливаются в `master` через Squash and Merge
- Итоговый коммит должен соответствовать формату Conventional Commits

# Архитектурная схема
https://drive.google.com/file/d/1pm5SVC891RtNJsOA02LdSOjeHyg8Us-Y/view?usp=sharing

## Порядок разработки

### 1. Подготовка окружения

#### macOS

```bash
go mod download
brew install golangci-lint    # опционально
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # опционально
```

#### Windows (PowerShell)

```bash
go mod download
choco install make
choco install golangci-lint    # опционально
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest   # опционально
```

#### Linux

```bash
go mod download
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 2. Запустить базу данных

```bash
make docker-up
```

### 3. Применить миграции

```bash
make migrate-up
```

### 4. Запустить систему начислений (accrual)

В отдельном терминале:

```bash
make accrual
```

### 5. Запустить приложение

В третьем терминале:

```bash
make run
```

---

## Команды Makefile

### Разработка

| Команда | Описание |
|---------|----------|
| `make build` | Собрать бинарный файл |
| `make run` | Запустить приложение локально |
| `make test` | Запустить все тесты |
| `make cover` | Запустить тесты с отчётом о покрытии |
| `make cover-html` | Открыть отчёт о покрытии в браузере |
| `make lint` | Проверить код линтером |
| `make clean` | Удалить бинарник и временные файлы |

### База данных

| Команда | Описание |
|---------|----------|
| `make docker-up` | Запустить PostgreSQL |
| `make docker-down` | Остановить PostgreSQL |
| `make docker-reset` | Удалить контейнер и все данные |
| `make docker-logs` | Посмотреть логи PostgreSQL |
| `make migrate-up` | Применить миграции |
| `make migrate-down` | Откатить миграции |
| `make migrate-status` | Статус миграций |

### Окружение

| Команда | Описание |
|---------|----------|
| `make dev` | Запустить БД и accrual (всё для разработки) |
| `make stop` | Остановить БД и accrual |
| `make accrual` | Запустить только accrual |

---

## Переменные окружения

Создайте файл `.env` в корне проекта (не коммитится):

```env
DATABASE_DSN=postgres://user:password@localhost:5432/loyalty?sslmode=disable
ACCRUAL_ADDR=http://localhost:8081
RUN_ADDRESS=localhost:8080
```

---

## Структура проекта

```
loyalty-service/
├── cmd/
│   ├── accrual/              # бинарники системы начислений
│   └── gophermart/           # точка входа приложения
│       └── main.go
├── internal/
│   ├── app/                  # инициализация приложения
│   ├── config/               # конфигурация (флаги, ENV)
│   ├── handler/              # HTTP-обработчики
│   │   ├── auth.go
│   │   ├── balance.go
│   │   ├── middleware.go
│   │   └── order.go
│   ├── logger/               # логирование
│   ├── model/                # модели данных
│   ├── repository/           # слой работы с БД
│   ├── service/              # бизнес-логика
│   └── worker/               # фоновые процессы
├── migrations/               # SQL-миграции
├── pkg/                      # переиспользуемые пакеты
│   └── luhn/                 # алгоритм Луна
├── .env.example              # шаблон переменных окружения
├── .gitignore
├── docker-compose.yml        # PostgreSQL + приложение
├── Dockerfile                # сборка приложения
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

```
users
- id (PK, int)
- login (string)
- password_hash (string)
- created_at (timestamp)
- updated_at (timestamp)
----------------------------
orders
- id (PK, int)
- number (string)
- user_id  (FK -> users.id)
- status (`NEW`, `INVALID`, `PROCESSING`, `PROCESSED`)
- accrual (decimal)
- uploaded_at (timestamp)
- updated_at (timestamp)
----------------------------
withdrawals
- id (PK, int)
- order_number  (string, не связан с таблицей orders)
- user_id  (FK -> users.id)
- sum (decimal)
- processed_at (timestamp)
```