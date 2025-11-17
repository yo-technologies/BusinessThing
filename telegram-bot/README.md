# Telegram Bot

Telegram бот для BusinessThing, написанный на Python.

## Функционал

- `/start` - приветствие с кнопкой для открытия веб-приложения
- `/help` - справка по командам

## Переменные окружения

- `TELEGRAM_BOT_TOKEN` - токен бота от BotFather (обязательный)
- `WEBAPP_URL` - URL веб-приложения (по умолчанию: https://businessthing.ru)

## Запуск локально

```bash
pip install -r requirements.txt
export TELEGRAM_BOT_TOKEN="your_bot_token"
export WEBAPP_URL="https://businessthing.ru"
python main.py
```

## Запуск в Docker

```bash
docker compose up telegram-bot
```

## Сборка образа

```bash
docker build -t telegram-bot .
```
