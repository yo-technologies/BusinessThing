"use client";

import { useMemo } from "react";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";

/**
 * Хук для получения роли пользователя в текущей организации
 */
export const useCurrentRole = () => {
  const { organizations, loading } = useAuth();
  const { currentOrg } = useOrganization({
    organizations,
    authLoading: loading,
  });

  const role = useMemo(() => {
    if (!currentOrg?.id) return null;

    const org = organizations.find((o) => o.id === currentOrg.id);

    return org?.role || null;
  }, [currentOrg?.id, organizations]);

  const isAdmin = role === "admin";

  return { role, isAdmin };
};
