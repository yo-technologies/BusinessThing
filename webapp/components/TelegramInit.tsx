"use client";

import { useEffect, useState } from "react";
import { init, miniApp, retrieveLaunchParams } from "@telegram-apps/sdk-react";

/**
 * Компонент для инициализации Telegram SDK.
 * Должен быть размещен в корне приложения.
 */
export function TelegramInit() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (typeof window === "undefined" || initialized) {
      return;
    }

    try {
      console.log("Initializing Telegram Mini App SDK...");
      
      // Получаем параметры запуска
      const params = retrieveLaunchParams();
      console.log("Launch params retrieved:", params);

      // Инициализируем все компоненты SDK
      init();
      
      // Сигнализируем, что приложение готово
      if (miniApp.mount.isAvailable()) {
        miniApp.ready();
        console.log("Telegram Mini App is ready");
      }

      // Разворачиваем viewport
      if (miniApp.mount.isAvailable()) {
        miniApp.setHeaderColor('#000000');
        miniApp.setBackgroundColor('#000000');
      }
      
      setInitialized(true);
    } catch (error) {
      console.error("Failed to initialize Telegram SDK:", error);
      // Даже если инициализация не удалась, отмечаем как выполненную,
      // чтобы не пытаться инициализировать повторно
      setInitialized(true);
    }
  }, [initialized]);

  return null;
}
