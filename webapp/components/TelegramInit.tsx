"use client";

import { useEffect, useState } from "react";
import { init, miniApp, retrieveLaunchParams, viewport } from "@telegram-apps/sdk-react";

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

      // startParam может быть в params.startParam или в params.tgWebAppData.start_param
      const startParam =
        params.startParam || (params as any).tgWebAppData?.start_param;

      // Проверяем и логируем приглашение
      if (startParam) {
        const startParamStr = String(startParam);

        console.log("StartParam as string:", startParamStr);
        if (startParamStr.startsWith("invitation_")) {
          const token = startParamStr.replace("invitation_", "");
        } else {
        }
      }

      // Инициализируем все компоненты SDK
      init();

      // Сигнализируем, что приложение готово
      if (miniApp.mount.isAvailable()) {
        miniApp.ready();
        console.log("Telegram Mini App is ready");
      }

      // Разворачиваем viewport и переходим в полноэкранный режим
      if (miniApp.mount.isAvailable()) {
        miniApp.setHeaderColor("#000000");
        miniApp.setBackgroundColor("#000000");
      }

      if (viewport.mount.isAvailable()) {
        viewport.mount();
        if (viewport.requestFullscreen.isAvailable()) {
          viewport.requestFullscreen();
        }
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
