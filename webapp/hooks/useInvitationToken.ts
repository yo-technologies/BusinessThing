"use client";

import { useEffect, useState } from "react";
import { retrieveLaunchParams } from "@telegram-apps/sdk-react";

const INVITATION_PROCESSED_KEY = "businessthing_invitation_processed";

/**
 * Хук для проверки наличия токена приглашения в startParam из Telegram или window.location
 * @returns true если есть токен приглашения, false в противном случае
 */
export const useHasInvitation = () => {
  const [hasInvitation, setHasInvitation] = useState<boolean>(false);

  useEffect(() => {
    if (typeof window === "undefined") return;

    // Проверяем, не было ли уже обработано приглашение в этой сессии
    const processed = sessionStorage.getItem(INVITATION_PROCESSED_KEY);

    if (processed) {
      console.log(
        "[useHasInvitation] Invitation already processed in this session",
      );
      setHasInvitation(false);

      return;
    }

    // Проверяем URL параметры
    const urlParams = new URLSearchParams(window.location.search);
    const urlToken = urlParams.get("token");

    if (urlToken) {
      setHasInvitation(true);

      return;
    }

    // Проверяем startParam из Telegram
    try {
      const params = retrieveLaunchParams();
      // startParam может быть в params.startParam или в params.tgWebAppData.start_param
      const startParam =
        params.startParam || (params as any).tgWebAppData?.start_param;

      if (startParam) {
        const startParamStr = String(startParam);

        // startParam может содержать токен приглашения в формате invitation_<token>
        if (startParamStr.startsWith("invitation_")) {
          console.log(
            "[useHasInvitation] Found invitation in startParam, setting hasInvitation=true",
          );
          setHasInvitation(true);
        } else {
          console.log("[useHasInvitation] Setting hasInvitation=false");
          setHasInvitation(false);
        }
      } else {
        setHasInvitation(false);
      }
    } catch (error) {
      setHasInvitation(false);
    }
  }, []);

  return hasInvitation;
};

/**
 * Функция для пометки приглашения как обработанного
 */
export const markInvitationAsProcessed = () => {
  if (typeof window !== "undefined") {
    console.log("[markInvitationAsProcessed] Marking invitation as processed");
    sessionStorage.setItem(INVITATION_PROCESSED_KEY, "true");
  }
};

/**
 * Функция для очистки флага обработки приглашения
 */
export const clearInvitationProcessed = () => {
  if (typeof window !== "undefined") {
    console.log(
      "[clearInvitationProcessed] Clearing invitation processed flag",
    );
    sessionStorage.removeItem(INVITATION_PROCESSED_KEY);
  }
};
