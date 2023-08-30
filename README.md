# [Тестовое задание для стажёра Backend](https://github.com/avito-tech/backend-trainee-assignment-2023)

### Выполнил Степанов Алексей Александрович

## PostgreSQL Query для создания таблиц:
```sql
CREATE TABLE user_segment_relations (
    user_id INTEGER,
    segment_id INTEGER,
    CONSTRAINT unique_user_segment UNIQUE (user_id, segment_id)
);

CREATE TABLE segments (
	id SERIAL UNIQUE,
	slug TEXT PRIMARY KEY
);
```

## Запуск:

Запуск dev-среды с помощью docker-compose:

```bash
docker-compose up
```

Запуск с помощью go run:

```bash
// Среда, в которой происходит запуск, должна иметь переменные окружения:
// DB_HOST
// DB_PORT
// DB_USER
// DB_PASSWORD
// DB_NAME
go run ./cmd/main.go // -create-tables=true - запуск с автоматическим созданием таблиц в БД
```

## Комментарии к решению:

Так как в задании не был указан механизм добавления новых пользователей в БД, мною было решено считать любой ID пользователя существующим.
Т.е. если в запросе на модификацию нового пользователя указан ID, который не существует в БД, то он будет добавлен в БД.

Также мною было решено требовать соответствия тела запроса с предполагаемым.
Т.е. если в запросе на модификацию нового пользователя указаны лишние поля, то запрос будет отклонён.
При этом, если некоторых полей не хватает, то им будут присвоены значения по умолчанию.

При удалении несуществующего сегмента, будет возвращён код StatusOK, но в БД ничего не изменится.

## Swagger
Swagger UI доступен по адресу: /swagger
> swagger.yaml и swagger.json находятся в папке [docs](./docs/)

## Примеры запросов:
// TODO
