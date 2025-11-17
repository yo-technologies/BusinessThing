"use client";

import { useMemo } from "react";

import { Api as AgentApi } from "@/api/api.agent.generated";
import { Api as CoreApi } from "@/api/api.core.generated";
import { getOrganizationsFromToken, Organization } from "@/utils/jwt";

const TOKEN_STORAGE_KEY = "businessthing_auth_token";

let authToken: string | null = null;

// Инициализация токена из localStorage при загрузке модуля
if (typeof window !== "undefined") {
  authToken = localStorage.getItem(TOKEN_STORAGE_KEY);
}

// Event emitter для обновления токена
const tokenUpdateListeners = new Set<() => void>();

export const onTokenUpdate = (listener: () => void) => {
  tokenUpdateListeners.add(listener);

  return () => tokenUpdateListeners.delete(listener);
};

const notifyTokenUpdate = () => {
  tokenUpdateListeners.forEach((listener) => listener());
};

export const setAuthToken = (token: string | null) => {
  authToken = token;
  if (typeof window !== "undefined") {
    if (token) {
      localStorage.setItem(TOKEN_STORAGE_KEY, token);
    } else {
      localStorage.removeItem(TOKEN_STORAGE_KEY);
    }
  }
  notifyTokenUpdate();
};

export const getAuthToken = () => authToken;

/**
 * Обновляет токен на сервере и возвращает новый токен с актуальным списком организаций
 */
export const refreshAuthToken = async (): Promise<{
  token: string;
  organizations: Organization[];
}> => {
  const coreApi = createCoreApi();
  const response = await coreApi.v1.authServiceRefreshToken();

  const newToken = response.data.accessToken || "";

  if (newToken) {
    setAuthToken(newToken);
  }

  const organizations = getOrganizationsFromToken(newToken);

  return {
    token: newToken,
    organizations,
  };
};

const createCoreApi = () => {
  const api = new CoreApi({
    baseURL: "https://core.businessthing.ru/api",
    securityWorker: () =>
      authToken
        ? {
            headers: {
              Authorization: `Bearer ${authToken}`,
            },
          }
        : {},
    secure: true,
  });

  // Добавляем интерсептор для обработки 401 ошибок
  const originalRequest = api.request.bind(api);

  api.request = async <T = any, E = any>(params: any): Promise<any> => {
    try {
      return await originalRequest<T, E>(params);
    } catch (error: any) {
      // Если получили 401, очищаем токен
      if (error?.status === 401 || error?.response?.status === 401) {
        console.warn("Received 401, clearing invalid token");
        setAuthToken(null);
      }
      throw error;
    }
  };

  return api;
};

const createAgentApi = () => {
  const api = new AgentApi({
    baseURL: "https://agent.businessthing.ru/api",
    securityWorker: () =>
      authToken
        ? {
            headers: {
              Authorization: `Bearer ${authToken}`,
            },
          }
        : {},
    secure: true,
  });

  // Добавляем интерсептор для обработки 401 ошибок
  const originalRequest = api.request.bind(api);

  api.request = async <T = any, E = any>(params: any): Promise<any> => {
    try {
      return await originalRequest<T, E>(params);
    } catch (error: any) {
      // Если получили 401, очищаем токен
      if (error?.status === 401 || error?.response?.status === 401) {
        console.warn("Received 401, clearing invalid token");
        setAuthToken(null);
      }
      throw error;
    }
  };

  return api;
};

export const useApiClients = () => {
  const core = useMemo(() => createCoreApi(), []);
  const agent = useMemo(() => createAgentApi(), []);

  return { core, agent };
};
