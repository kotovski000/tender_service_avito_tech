# Tender-service


## Установка

```
git clone https://github.com/ваш-репозиторий/tender_service_avito_tech.git
cd tender_service_avito_tech
cd tender
go mod tidy
```

## Настройка
Приложение использует параметры окружения:

SERVER_ADDRESS — адрес и порт, который будет слушать HTTP сервер при запуске.

POSTGRES_USER — имя пользователя для подключения к PostgreSQL.

POSTGRES_PASSWORD — пароль для подключения к PostgreSQL.

POSTGRES_HOST — хост для подключения к PostgreSQL.

POSTGRES_PORT — порт для подключения к PostgreSQL.

POSTGRES_DB — имя базы данных PostgreSQL, которую будет использовать приложение.


## Структура проекта
- задание/: В папке "задание" размещена задача.
- cmd/main.go: Главный файл приложения, точка входа сервера.
- db/: Пакет для инициализации базы данных.
- handlers/: Пакет с обработчиками API запросов.
- models/:В директории находятся структуры данных для работы.

## Запуск приложения

docker compose up -d
