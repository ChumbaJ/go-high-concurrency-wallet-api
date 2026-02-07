# Wallet API

HTTP API для работы с кошельками: пополнение, списание, получение баланса.  
Go + PostgreSQL + очередь с батчингом для высокой пропускной способности.

---

## Быстрый старт

### 1. Клонировать и настроить

```bash
git clone <repo-url> wallet-api
cd wallet-api
```

Создайте файл `config.env` и скопируйте в него содержимое `config.env.example` (Ctrl+C → Ctrl+V).  
При необходимости измените `DB_PASSWORD` и другие параметры.

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

**Готово.** API доступен на `http://localhost:8080`.

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
