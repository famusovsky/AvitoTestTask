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
)
```

## Комментарии к решению:

Так как в задании не был указан механизм добавления новых пользователей в БД, мною было решено считать любой ID пользователя существующим.
Т.е. если в запросе на модификацию нового пользователя указан ID, который не существует в БД, то он будет добавлен в БД.

Также мною было решено требовать полное соответствие тела запроса с предполагаемым.
Т.е. если в запросе на модификацию нового пользователя не указаны все поля или указаны лишние поля, то запрос будет отклонён.