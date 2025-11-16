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

interface AuthState {
  isAuthenticated: boolean;
  loading: boolean;
  user: AuthUser | null;
  isNewUser: boolean;
  organizations: Organization[];
}

export const useAuth = (): AuthState => {
  const initData = typeof window !== "undefined" ? useRawInitData() : null;
  const { core } = useApiClients();
  const [state, setState] = useState<AuthState>(() => {
    // Проверяем наличие токена при инициализации
    const existingToken = getAuthToken();
    if (existingToken) {
      const organizations = getOrganizationsFromToken(existingToken);
      return {
        isAuthenticated: true,
        loading: true, // всё ещё нужно проверить через authenticate
        user: null,
        isNewUser: false,
        organizations,
      };
    }
    return {
      isAuthenticated: false,
      loading: true,
      user: null,
      isNewUser: false,
      organizations: [],
    };
  });

  const authenticate = async () => {
    if (typeof window === "undefined") {
      console.log("useAuth: SSR mode, skipping auth");
      return;
    }
    
    // Если токен уже есть, попробуем его обновить без повторной аутентификации
    const existingToken = getAuthToken();
    if (existingToken) {
      try {
        const response = await core.v1.authServiceRefreshToken();
        const { accessToken } = response.data;
        
        if (accessToken) {
          setAuthToken(accessToken);
          const organizations = getOrganizationsFromToken(accessToken);
          const userId = getUserIdFromToken(accessToken);
          let user = null;

          if (userId) {
            const userResponse = await core.v1.userServiceGetUser(userId);
            user = userResponse.data.user ?? null;
          }

          setState((prev) => ({
            ...prev,
            isAuthenticated: true,
            loading: false,
            organizations,
            user,
          }));
        }
        return;
      } catch (error) {
        console.warn("Token refresh failed, will try to authenticate with Telegram", error);
        setAuthToken(null);
      }
    }
    
    if (!initData) {
      console.warn("useAuth: initData is not available. Make sure the app is opened from Telegram.");
      setState((prev) => ({ ...prev, loading: false }));
      return;
    }

    try {
      console.log("Authenticating with initData:", initData);
      
      const payload: CoreAuthenticateWithTelegramRequest = { initData };

      const response = await core.v1.authServiceAuthenticateWithTelegram(payload, {
        secure: false,
      });

      const { accessToken, user, isNewUser } = response.data;

      let organizations: Organization[] = [];

      if (accessToken) {
        setAuthToken(accessToken);
        organizations = getOrganizationsFromToken(accessToken);
      }

      setState((prev) => ({
        ...prev,
        isAuthenticated: Boolean(accessToken),
        loading: false,
        user: user ?? null,
        isNewUser: Boolean(isNewUser),
        organizations,
      }));
    } catch (error) {
      console.error("Auth error", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  useEffect(() => {
    let cancelled = false;

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
        setState((prev) => ({
          ...prev,
          organizations,
          isAuthenticated: true,
        }));
      }
    });

    return () => {
      unsubscribe();
    };
  }, []);

  return state;
};
