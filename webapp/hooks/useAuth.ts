// alpha/webapp/hooks/useAuth.ts
"use client";

import { useEffect, useState } from "react";
import { useRawInitData } from "@telegram-apps/sdk-react";

import { CoreAuthenticateWithTelegramRequest } from "@/api/api.core.generated";
import { setAuthToken, useApiClients, onTokenUpdate, getAuthToken } from "@/api/client";
import { getOrganizationsFromToken, Organization } from "@/utils/jwt";

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
      // Проверяем срок действия токена
      const payload = require('@/utils/jwt').decodeJWT(existingToken);
      const isExpired = payload ? payload.exp * 1000 < Date.now() : true;
      
      if (!isExpired && payload) {
        const organizations = getOrganizationsFromToken(existingToken);
        const userInfo = { userId: payload.sub, role: organizations[0]?.role || '' };
        return {
          isAuthenticated: true,
          loading: false,
          user: null,
          isNewUser: false,
          organizations,
          userInfo,
        };
      } else {
        // Токен истек, очищаем его
        setAuthToken(null);
      }
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

    // Если токен уже есть и не истек, не нужно заново аутентифицироваться
    const existingToken = getAuthToken();
    if (existingToken) {
      const payload = require('@/utils/jwt').decodeJWT(existingToken);
      const isExpired = payload ? payload.exp * 1000 < Date.now() : true;
      
      if (!isExpired) {
        console.log("useAuth: Token already exists and valid, skipping authentication");
        return;
      } else {
        console.log("useAuth: Token expired, re-authenticating");
        setAuthToken(null);
      }
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
      } else {
        // Токен был очищен (например, из-за 401), переходим в состояние загрузки и пытаемся заново аутентифицироваться
        setState((prev) => ({
          ...prev,
          isAuthenticated: false,
          loading: true,
          organizations: [],
          userInfo: null,
        }));
        
        // Пытаемся заново аутентифицироваться
        if (initData) {
          authenticate().catch((err) => {
            console.error("Re-authentication failed", err);
            setState((prev) => ({ ...prev, loading: false }));
          });
        } else {
          setState((prev) => ({ ...prev, loading: false }));
        }
      }
    });

    return () => {
      unsubscribe();
    };
  }, [initData]);

  return state;
};
