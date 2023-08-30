# [Тестовое задание для стажёра Backend](https://github.com/avito-tech/backend-trainee-assignment-2023)

### Выполнил Степанов Алексей Александрович

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

## PostgreSQL Query для создания таблиц в БД:
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

GET /users/99 => 200 `[{"slug":"test1"},{"slug":"test2"}]`
> Возвращает список сегментов, в которые входит пользователь с id = 99, ("test1" и "test2") в формате JSON.

PATCH /users `{"id":10,"append":["test1","test2"],"remove":["test3","test4"]}` => 200 `"OK"`
> Добавляет пользователя с id = 10 в сегменты test1 и test2, а также удаляет его из сегментов test3 и test4.

POST /segments `{"slug":"test"}` => 200 `"OK"`
> Добавляет сегмент с именем test в БД.

DELETE /segments `{"slug":"test"}` => 200 `"OK"`
> Удаляет сегмент с именем test из БД.

