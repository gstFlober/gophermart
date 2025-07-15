Микросервис для управления баллами лояльности, обработки заказов и финансовых операций. Интегрируется с внешним сервисом начислений (Accrual).

## Особенности

- ✅ Регистрация и аутентификация пользователей
- ✅ Загрузка и отслеживание статуса заказов
- ✅ Управление балансом (просмотр/списание)
- ✅ История операций
- ✅ JWT-аутентификация
- ✅ Фоновая обработка заказов
- ✅ Подробное логирование
- ✅ Graceful shutdown

## Технологический стек

- **Язык:** Go 1.21+
- **Фреймворк:** Echo v4
- **База данных:** PostgreSQL
- **Логирование:** Zerolog
- **Аутентификация:** JWT

## Конфигурация

Создайте файл `config.yaml` в корне проекта:

```yaml
server:
  address: ":8080"               # Порт сервера
  shutdown_timeout: 30s          # Таймаут graceful shutdown
  read_timeout: 15s              # Таймаут чтения запросов
  write_timeout: 15s             # Таймаут записи ответов

auth:
  jwt_secret: "supersecretkey"   # Секрет для подписи JWT
  jwt_access_expiry: 15m         # Время жизни access токена
  jwt_refresh_expiry: 168h       # Время жизни refresh токена (7 дней)
  token_refresh_leeway: 5m       # Допустимое время обновления токена

database:
  type: "postgres"               # Тип БД
  connection_retries: 5          # Попытки подключения к БД
  retry_delay: 1s                # Задержка между попытками
  postgres:
    host: "postgresql"           # Хост БД
    port: "5432"                 # Порт БД
    uri: "postgresql://..."      # URI подключения
    user: "postgres"             # Пользователь БД
    password: "postgres"         # Пароль БД
    dbname: "praktikum"          # Имя БД
    sslmode: "disable"           # Режим SSL
    max_open_conns: 30           # Макс. открытых соединений
    max_idle_conns: 10           # Макс. неактивных соединений
    max_lifetime: 30m            # Время жизни соединения

logger:
  level: "debug"                 # Уровень логирования (debug, info, warn, error)
  format: "text"                 # Формат логов (text, json)
  output: "stdout"               # Вывод логов (stdout, file)
  with_caller: true              # Показывать место вызова

accural: "http://localhost:9099" # Адрес сервиса начислений