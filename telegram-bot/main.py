#!/usr/bin/env python3
import logging
import os
import signal
import sys
from contextlib import asynccontextmanager

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup, WebAppInfo
from telegram.ext import Application, CommandHandler, ContextTypes

# Configure logging
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)
logger = logging.getLogger(__name__)


class TelegramBot:
    def __init__(self, token: str, webapp_url: str):
        self.token = token
        self.webapp_url = webapp_url
        self.application = None

    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        """Handle /start command"""
        user = update.effective_user
        text = (
            f"Привет, {user.first_name}! 👋\n\n"
            "Добро пожаловать в BusinessThing — твой личный бизнес-ассистент.\n\n"
            "Нажми на кнопку ниже, чтобы открыть приложение:"
        )

        keyboard = [
            [InlineKeyboardButton("🚀 Открыть приложение", web_app=WebAppInfo(url=self.webapp_url))]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)

        await update.message.reply_text(text, reply_markup=reply_markup)

    async def help_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        """Handle /help command"""
        text = (
            "Команды бота:\n\n"
            "/start - Открыть приложение\n"
            "/help - Показать эту справку"
        )
        await update.message.reply_text(text)

    async def unknown_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        """Handle unknown commands"""
        text = "Используй /start, чтобы открыть приложение."
        await update.message.reply_text(text)

    def setup_handlers(self):
        """Setup command handlers"""
        self.application.add_handler(CommandHandler("start", self.start_command))
        self.application.add_handler(CommandHandler("help", self.help_command))

    async def run(self):
        """Run the bot"""
        logger.info("Starting Telegram bot...")
        
        self.application = Application.builder().token(self.token).build()
        self.setup_handlers()

        bot_info = await self.application.bot.get_me()
        logger.info(f"Authorized as @{bot_info.username}")
        logger.info("Bot started. Waiting for updates...")

        await self.application.initialize()
        await self.application.start()
        await self.application.updater.start_polling(drop_pending_updates=True)

    async def stop(self):
        """Stop the bot gracefully"""
        if self.application:
            logger.info("Shutting down bot...")
            await self.application.updater.stop()
            await self.application.stop()
            await self.application.shutdown()
            logger.info("Bot stopped")


def main():
    """Main function"""
    bot_token = os.getenv("TELEGRAM_BOT_TOKEN")
    webapp_url = os.getenv("WEBAPP_URL")

    if not bot_token:
        logger.error("TELEGRAM_BOT_TOKEN environment variable is required")
        sys.exit(1)
    
    if not webapp_url:
        logger.error("WEBAPP_URL environment variable is required")
        sys.exit(1)

    bot = TelegramBot(token=bot_token, webapp_url=webapp_url)

    # Setup signal handlers
    import asyncio
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    def signal_handler(sig, frame):
        logger.info(f"Received signal {sig}")
        loop.create_task(bot.stop())
        loop.stop()

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        loop.run_until_complete(bot.run())
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        loop.run_until_complete(bot.stop())
        loop.close()


if __name__ == "__main__":
    main()
