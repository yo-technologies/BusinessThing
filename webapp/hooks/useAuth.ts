// alpha/webapp/hooks/useAuth.ts
"use client";

import { useEffect, useState } from "react";
import { useRawInitData } from "@telegram-apps/sdk-react";

import { CoreAuthenticateWithTelegramRequest } from "@/api/api.core.generated";
import { setAuthToken, useApiClients } from "@/api/client";
import { getOrganizationsFromToken, Organization } from "@/utils/jwt";

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
  reAuthenticate: () => Promise<void>;
}

export const useAuth = (): AuthState => {
  const initData = typeof window !== "undefined" ? useRawInitData() : null;
  const { core } = useApiClients();
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    loading: true,
    user: null,
    isNewUser: false,
    organizations: [],
    reAuthenticate: async () => {},
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

  // Добавляем функцию reAuthenticate в state
  useEffect(() => {
    setState((prev) => ({
      ...prev,
      reAuthenticate: authenticate,
    }));
  }, [initData]);

  return state;
};
