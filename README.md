# Subscription Aggregation Service

REST-сервис на Go для учета онлайн-подписок пользователей и подсчета суммарной стоимости подписок за выбранный период.

## Стек

- Go 1.23
- PostgreSQL 16
- chi router
- pgx
- goose migrations
- zap logger
- Docker Compose
- OpenAPI 3.0

## Запуск

```bash
cp .env.example .env
docker compose up --build
```

API будет доступен на `http://localhost:8080`.

Swagger/OpenAPI YAML:

```text
http://localhost:8080/swagger/openapi.yaml
```

Health check:

```bash
curl http://localhost:8080/health
```

## Модель подписки

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

`end_date` опциональна. Даты передаются в формате `MM-YYYY`.

## CRUDL

### Создать подписку

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Получить подписку

```bash
curl http://localhost:8080/api/v1/subscriptions/1
```

### Получить список

```bash
curl 'http://localhost:8080/api/v1/subscriptions?limit=20&offset=0'
```

С фильтрами:

```bash
curl 'http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus'
```

### Обновить подписку

```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/1 \
  -H 'Content-Type: application/json' \
  -d '{"price": 500, "end_date": "12-2025"}'
```

Чтобы убрать дату окончания, передайте пустую строку:

```json
{ "end_date": "" }
```

### Удалить подписку

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/1
```

Удаление мягкое: запись помечается `deleted_at`.

## Подсчет суммы за период

```bash
curl 'http://localhost:8080/api/v1/subscriptions/sum?from=07-2025&to=09-2025'
```

С фильтрами:

```bash
curl 'http://localhost:8080/api/v1/subscriptions/sum?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus'
```

Сумма считается помесячно: если подписка активна в месяце из выбранного периода, к сумме добавляется ее месячная цена. Например, подписка за 400 рублей с `start_date=07-2025` без `end_date` даст `1200` за период `07-2025` — `09-2025`.

## Миграции

В Docker Compose миграции выполняются автоматически сервисом `migrate`.

Локально можно выполнить:

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" up
```

## Конфигурация

Конфиг можно задавать через `.env` или YAML-файл.

Переменные окружения имеют приоритет над YAML.

Для YAML укажите:

```bash
CONFIG_PATH=config.yaml
```

