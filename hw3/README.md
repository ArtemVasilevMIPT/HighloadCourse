## Задание №3 по курсу HighLoad
### Зависимости
* RabbitMQ
* SQLite3
### Сборка
```
$ go build -o bin/server cmd/app/main.go
$ go build -o bin/email cmd/email/main.go
```
### Перед запуском
1. Запустить `scripts/create_db.sql`, чтобы создать базу данных с пользователями
2. Запустить `scripts/create_config.sh`, чтобы создать файл конфигурации для почтового сервиса
3. Отредактировать файл конфигурации из папки `config`:
   1. SMTP_HOSTNAME - имя smtp сервера, например `smtp.yandex.ru`
   2. SMTP_PORT - порт smtp сервера
   3. SMTP_USERNAME - имя пользователя, например `example@example.com`
   4. SMTP_PASSWORD - пароль для авторизации на сервере
   
### Запуск
Почтовый сервис: `./bin/email`

Основное приложение: `./bin/server`

* `localhost:8080/login` - страница входа
* `localhost:8080/register` - страница регистрации
* `localhost:8080/reset-password` - страница для смены пароля

