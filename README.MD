
# SalesTracker

## Описание

SalesTracker — сервис для управления записями и аналитики данных

## Состав репозитория

- **cmd/SalesTracker/main.go** — точка входа через FX DI.
- **internal/**
  - **app/analytics** — бизнес-логика работы с аналитикой.
  - **app/transactions** — бизнес-логика транзакций.
  - **config/** — загрузка конфигурации из YAML.
  - **di/** — реализация зависимостей через UberFX.
  - **domain/analytic** — модель аналитики
  - **domain/transaction** — модель транзакции
  - **storage/postgres** — работа с PostgreSQL (CRUD).
  - **web/** — HTTP-обработчики и роутер.
- **config/local.yaml** — пример конфигурации.
- **migrations/** — SQL-миграции для PostgreSQL.
- **docs/** — Swagger-документация.
- **web/index.html** — простая страница создания транзакций .
- **docker-compose.yml** — запуск PostgreSQL через Docker.
- **.env.example** - пример env файла для кредов.


---

## Быстрый старт

### 1. Запуск инфраструктуры

```sh
docker-compose up -d
```
(Запустит контейнеры: postgres → порт 5433)

### 2. Настроить переменные окружения и конфигурацию
(пример в .env.example + config/local.yaml)

### 3. Применить миграции (migrate):

```sh
migrate -path migrations -database "postgres://user:password@localhost:5433/dbname?sslmode=disable" up
```

### 4. Запуск сервиса

```sh
go run ./cmd/salestracker/main.go
```

Сервис стартует на порту 8080.

## API

- **POST /items** — создание транзакции;
- **GET /items** — получение списка транзакций;
- **GET /items/{id}** — получение информации о транзакции по ID;
- **PUT /items/{id}** — изменение информации о транзакции по ID;
- **DELETE /items/{id}** — удаление транзакции по ID;
- **GET /items/export** — экспорт транзакций в CSV;

- **GET /analytics** — получение аналитики по транзакциям;
- **GET /analytics/export** —  экспорт аналитики в CSV;
- **Swagger**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## Веб-интерфейс
Откройте index.html в браузере — рабочий пример создания транзакций и получения аналитики.


## Тесты
Юнит-тесты: `go test ./internal/...`

## Миграции

- `migrations/000001_create_transaction_table.up.sql` — создание таблиц.
- `migrations/000001_create_transaction_table.down.sql` — удаление таблиц.

---

## Логирование и метрики
Логирование реализовано через wbf/zlog (используется в internal/*).

## Зависимости

- Go 1.25+
- PostgreSQL 16+
- Docker (для локального запуска инфраструктуры)

---

## Swagger

- Swagger: [docs/swagger.yaml](docs/swagger.yaml)
- Документация генерируется автоматически и доступна по `/swagger/*`.