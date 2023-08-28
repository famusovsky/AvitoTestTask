# [Тестовое задание для стажёра Backend](https://github.com/avito-tech/backend-trainee-assignment-2023)

## Выполнил Степанов Алексей Александрович

PostgreSQL Query для создания таблиц:
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