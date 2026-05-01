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

# Структура каталогов

```
loyalty-service/
├── cmd/
│   └── gofermart/
│       └── main.go              # точка входа в приложение
├── internal/
│   ├── app/                     # сборка приложения
│   │   └── app.go
│   ├── config/                  # конфигурация
│   │   └── config.go
│   ├── handler/                 # HTTP-хендлеры
│   │   ├── auth.go
│   │   ├── order.go
│   │   ├── balance.go
│   │   └── middleware.go
│   ├── model/                   # модели данных
│   │   ├── user.go
│   │   ├── order.go
│   │   └── withdrawal.go
│   ├── repository/              # слой работы с БД
│   │   ├── pg.go
│   │   ├── user.go
│   │   ├── order.go
│   │   ├── balance.go
│   │   └── withdrawal.go
│   ├── service/                 # бизнес-логика
│   │   ├── auth.go
│   │   ├── order.go
│   │   └── balance.go
│   └── worker/                  # фоновые процессы
│       └── accrual.go
├── migrations/                  # миграции на верхнем уровне
│   ├── 000001_init.up.sql
│   ├── 000001_init.down.sql
│   ├── 000002_add_withdrawals.up.sql
│   └── 000002_add_withdrawals.down.sql
├── pkg/
│   └── luhn/
│       └── luhn.go
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── README.md
```