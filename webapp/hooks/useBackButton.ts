"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { backButton } from "@telegram-apps/sdk-react";

/**
 * Хук для управления Telegram BackButton
 * @param show - показывать ли кнопку
 * @param onClick - кастомный обработчик клика (опционально)
 */
export const useBackButton = (show: boolean = true, onClick?: () => void) => {
  const router = useRouter();

  useEffect(() => {
    if (typeof window === "undefined") return;

    try {
      // Монтируем BackButton если доступен
      if (!backButton.mount.isAvailable()) {
        console.warn("BackButton is not available");

        return;
      }

      if (!backButton.isMounted()) {
        backButton.mount();
      }

      if (show) {
        backButton.show();

        const handleClick = () => {
          if (onClick) {
            onClick();
          } else {
            router.back();
          }
        };

        backButton.onClick(handleClick);

        return () => {
          backButton.offClick(handleClick);
          backButton.hide();
        };
      } else {
        backButton.hide();
      }
    } catch (error) {
      console.error("Failed to setup BackButton:", error);
    }
  }, [show, onClick, router]);
};
