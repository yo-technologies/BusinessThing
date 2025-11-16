// alpha/webapp/hooks/useAuth.ts
"use client";

import { useEffect, useState } from "react";
import { useRawInitData } from "@telegram-apps/sdk-react";

import { CoreAuthenticateWithTelegramRequest } from "@/api/api.core.generated";
import { setAuthToken, useApiClients, onTokenUpdate, getAuthToken } from "@/api/client";
import { getOrganizationsFromToken, getUserIdFromToken, Organization } from "@/utils/jwt"; // Import getUserIdFromToken

interface AuthUser {
  id?: string;
  firstName?: string;
  lastName?: string;
}

interface UserInfo {
  userId: string;
  role: string;
}

interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
  user: AuthUser | null;
  isNewUser: boolean;
  organizations: Organization[];
  userInfo: UserInfo | null;
}

export const useAuth = (): AuthState => {
  const initData = typeof window !== "undefined" ? useRawInitData() : null;
  const { core } = useApiClients();
  const [state, setState] = useState<AuthState>(() => {
    // Проверяем наличие токена при инициализации
    const existingToken = getAuthToken();
    if (existingToken) {
      const organizations = getOrganizationsFromToken(existingToken);
      const payload = existingToken ? require('@/utils/jwt').decodeJWT(existingToken) : null;
      const userInfo = payload ? { userId: payload.sub, role: organizations[0]?.role || '' } : null;
      return {
        isAuthenticated: true,
        loading: false, // если токен есть, считаем что загрузка завершена
        user: null,
        isNewUser: false,
        organizations,
        userInfo,
      };
    }
    return {
      isAuthenticated: false,
      loading: true,
      user: null,
      isNewUser: false,
      organizations: [],
      userInfo: null,
    };
  });

  const authenticate = async () => {
    if (typeof window === "undefined") {
      console.log("useAuth: SSR mode, skipping auth");
      return;
    }
    
    if (!initData) {
      console.warn("useAuth: initData is not available. Make sure the app is opened from Telegram.");
      setState((prev) => ({ ...prev, loading: false }));
      return;
    }

    try {
      console.log("Authenticating with initData:", initData);
      
      const requestPayload: CoreAuthenticateWithTelegramRequest = { initData };

      const response = await core.v1.authServiceAuthenticateWithTelegram(requestPayload, {
        secure: false,
      });

      const { accessToken, user, isNewUser } = response.data;

      let organizations: Organization[] = [];

      if (accessToken) {
        setAuthToken(accessToken);
        organizations = getOrganizationsFromToken(accessToken);
      }

      const payload = accessToken ? require('@/utils/jwt').decodeJWT(accessToken) : null;
      const userInfo = payload ? { userId: payload.sub, role: organizations[0]?.role || '' } : null;

      setState((prev) => ({
        ...prev,
        isAuthenticated: Boolean(accessToken),
        loading: false,
        user: user ?? null,
        isNewUser: Boolean(isNewUser),
        organizations,
        userInfo,
      }));
    } catch (error) {
      console.error("Auth error", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  useEffect(() => {
    let cancelled = false;

    // Если токен уже есть, не нужно заново аутентифицироваться
    const existingToken = getAuthToken();
    if (existingToken) {
      console.log("useAuth: Token already exists, skipping authentication");
      return;
    }

    authenticate().catch((err) => {
      if (!cancelled) {
        console.error("Auth failed", err);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [initData]);

  // Подписываемся на обновления токена
  useEffect(() => {
    const unsubscribe = onTokenUpdate(() => {
      const token = getAuthToken();
      if (token) {
        const organizations = getOrganizationsFromToken(token);
        const payload = require('@/utils/jwt').decodeJWT(token);
        const userInfo = payload ? { userId: payload.sub, role: organizations[0]?.role || '' } : null;
        setState((prev) => ({
          ...prev,
          organizations,
          isAuthenticated: true,
          userInfo,
        }));
      }
    });

    return () => {
      unsubscribe();
    };
  }, []);

  return state;
};
