# PR Reviewer Assignment Service

Микросервис для автоназначения ревьюеров на пулл реквесты и управления командами

### Требования 
    Docker
     

### Запуск с Docker Compose 

```
git clone https://github.com/forzeyy/avito-autumn.git
cd avito-autumn

docker-compose up --build
```

Сервис будет доступен по адресу: http://localhost:8080

БД и миграции инициализируются при запуске.

## Пример .env
Хранится в cmd/avito
```
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=avito
```