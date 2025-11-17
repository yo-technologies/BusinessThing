"use client";

import { useEffect, useState } from "react";
import { retrieveLaunchParams } from "@telegram-apps/sdk-react";

/**
 * Хук для проверки наличия токена приглашения в startParam из Telegram или window.location
 * @returns true если есть токен приглашения, false в противном случае
 */
export const useHasInvitation = () => {
  const [hasInvitation, setHasInvitation] = useState<boolean>(false);

  useEffect(() => {
    if (typeof window === "undefined") return;

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
