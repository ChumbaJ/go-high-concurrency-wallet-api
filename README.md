# Wallet API

HTTP API для работы с кошельками: пополнение, списание, получение баланса.  
Go + PostgreSQL + очередь с батчингом для высокой пропускной способности.

---

## Быстрый старт

### 1. Клонировать и настроить

```bash
git clone <repo-url> wallet-api
cd wallet-api
cp config.env.example config.env
# при необходимости отредактировать config.env (пароль БД и т.д.)
```

### 2. Запуск

**Docker Compose** (рекомендуется):

```bash
docker compose up -d
```

PostgreSQL + миграции + приложение. API доступен на `http://localhost:8080`.

**Без Docker** (локально):

```bash
# PostgreSQL должен быть запущен
go run ./cmd
```

---

## API

| Метод | URL                    | Описание                                                                                                        |
| ----- | ---------------------- | --------------------------------------------------------------------------------------------------------------- |
| POST  | `/api/v1/wallet`       | Пополнение или списание. Body: `{ "walletId": "uuid", "operationType": "DEPOSIT"\|"WITHDRAW", "amount": 1000 }` |
| GET   | `/api/v1/wallets/{id}` | Получить баланс кошелька                                                                                        |

---

## Load Test

### hey (CLI)

Установка:

```bash
go install github.com/rakyll/hey@latest
```

**Примеры:**

```bash
# GET баланса — 100 RPS, 10 с
hey -z 10s -q 100 -c 20 http://localhost:8080/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000

# POST пополнения — 500 RPS, 15 с
hey -z 15s -q 500 -c 50 -m POST -H "Content-Type: application/json" \
  -d '{"walletId":"550e8400-e29b-41d4-a716-446655440001","operationType":"DEPOSIT","amount":1000}' \
  http://localhost:8080/api/v1/wallet
```

Параметры `hey`:

- `-z` — длительность
- `-q` — запросов в секунду (RPS)
- `-c` — параллельные соединения

### load-test.js

Скрипт запускает 3 потока (2× DEPOSIT, 1× GET) суммарно ~2000 RPS:

```bash
node load-test.js
```

Переменные окружения: `BASE_URL` (по умолчанию `http://localhost:8080`), `DURATION` (по умолчанию `15`).
