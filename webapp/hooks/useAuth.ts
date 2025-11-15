// alpha/webapp/hooks/useAuth.ts
"use client";

import { useEffect, useState } from "react";
import { useRawInitData } from "@telegram-apps/sdk-react";

import { CoreAuthenticateWithTelegramRequest } from "@/api/api.core.generated";
import { setAuthToken, useApiClients } from "@/api/client";

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
}

export const useAuth = (): AuthState => {
  const initData = typeof window !== "undefined" ? useRawInitData() : null;
  const { core } = useApiClients();
  const [state, setState] = useState<AuthState>({
    isAuthenticated: false,
    loading: true,
    user: null,
    isNewUser: false,
  });

  useEffect(() => {
    if (typeof window === "undefined" || !initData) {
      return;
    }

    let cancelled = false;

    const authenticate = async () => {
      try {
        console.log("Authenticating with initData:", initData);
        
        const payload: CoreAuthenticateWithTelegramRequest = { initData };

        const response = await core.v1.authServiceAuthenticateWithTelegram(payload, {
          secure: false,
        });

        if (cancelled) {
          return;
        }

        const { accessToken, user, isNewUser } = response.data;

        if (accessToken) {
          setAuthToken(accessToken);
        }

        setState({
          isAuthenticated: Boolean(accessToken),
          loading: false,
          user: user ?? null,
          isNewUser: Boolean(isNewUser),
        });
      } catch (error) {
        console.error("Auth error", error);
        if (!cancelled) {
          setState((prev) => ({ ...prev, loading: false }));
        }
      }
    };

    authenticate();

    return () => {
      cancelled = true;
    };
  }, [initData]);

  return state;
};
