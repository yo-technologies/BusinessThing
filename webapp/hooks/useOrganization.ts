"use client";

import { useCallback, useEffect, useState } from "react";

import { Organization } from "@/utils/jwt";

const ORG_STORAGE_KEY = "businessthing_current_org_id";

interface OrganizationState {
  currentOrg: Organization | null;
  organizations: Organization[];
  loading: boolean;
  needsOrganization: boolean;
}

interface UseOrganizationProps {
  organizations: Organization[];
  authLoading?: boolean; // Добавляем флаг загрузки авторизации
}

export const useOrganization = ({ organizations, authLoading = false }: UseOrganizationProps) => {
  const [state, setState] = useState<OrganizationState>({
    currentOrg: null,
    organizations: [],
    loading: true,
    needsOrganization: false,
  });

  useEffect(() => {
    // Если авторизация еще не завершилась, ждем
    if (authLoading) {
      return;
    }

    if (organizations.length === 0) {
      setState({
        currentOrg: null,
        organizations: [],
        loading: false,
        needsOrganization: true,
      });
      return;
    }

    // Пытаемся взять из localStorage
    const storedOrgId = typeof window !== "undefined" ? localStorage.getItem(ORG_STORAGE_KEY) : null;
    let currentOrg = storedOrgId
      ? organizations.find((org) => org.id === storedOrgId) ?? null
      : null;

    // Если не нашли или нет в storage, берём первую
    if (!currentOrg) {
      currentOrg = organizations[0];
      if (currentOrg.id && typeof window !== "undefined") {
        localStorage.setItem(ORG_STORAGE_KEY, currentOrg.id);
      }
    }

    setState({
      currentOrg,
      organizations,
      loading: false,
      needsOrganization: false,
    });
  }, [organizations, authLoading]);

  const switchOrganization = useCallback(
    (orgId: string) => {
      const org = organizations.find((o) => o.id === orgId);
      if (org) {
        setState((prev) => ({ ...prev, currentOrg: org }));
        if (typeof window !== "undefined") {
          localStorage.setItem(ORG_STORAGE_KEY, orgId);
        }
      }
    },
    [organizations],
  );

  return {
    ...state,
    switchOrganization,
  };
};
