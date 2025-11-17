"use client";

import { useEffect, useState } from "react";
import { viewport } from "@telegram-apps/sdk-react";

/**
 * Хук для получения информации о viewport Telegram Mini App
 */
export function useTelegramViewport() {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    try {
      // Проверяем, доступен ли viewport
      if (viewport.mount.isAvailable()) {
        viewport.mount();
        setIsMounted(true);

        // Получаем текущее состояние fullscreen
        const currentState = viewport.isFullscreen();

        setIsFullscreen(currentState);

        // Подписываемся на изменения fullscreen
        const unsubscribe = viewport.isFullscreen.sub((value: boolean) => {
          setIsFullscreen(value);
        });

        return () => {
          unsubscribe();
        };
      }
    } catch (error) {
      console.warn("Failed to mount viewport:", error);
    }
  }, []);

  return { isFullscreen, isMounted };
}
