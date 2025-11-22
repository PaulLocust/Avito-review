# 📝 Avito-review - Backend Service for PR assignments

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-24+-2496ED?logo=docker)
![Swagger](https://img.shields.io/badge/Swagger-3.0-85EA2D?logo=swagger)

Backend-сервис, в котором можно назначать ревьюеров для Pull Request’ов

## 🚀 Как запускать

```bash
# 1. Клонируйте репозиторий
git clone https://github.com/PaulLocust/Avito-review.git

# 2. Запустите сервисы через Docker
docker-compose up -d
```

## После запуска доступны:
- 📚 http://localhost:8080/swagger API Documentation - место где можно поиграться с приложением

## 🛠 Технологии
- **Язык**: Go
- **База данных**: PostgreSQL
- **Инфраструктура**: Docker
- **Документация**: Swagger, OpenAPI
- **Архитектура**: Clean Architecture
- **Прочее**: Миграции, автогенерация DTO, логирование на каждом архитектурном слое
