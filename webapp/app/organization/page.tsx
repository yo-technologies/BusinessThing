"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import {
  BookOpenIcon,
  DocumentTextIcon,
  BriefcaseIcon,
  UsersIcon,
  Cog6ToothIcon,
  ArrowRightIcon,
  BuildingOfficeIcon,
  LightBulbIcon,
} from "@heroicons/react/24/outline";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { CoreOrganization } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";

const sections = [
  {
    key: "users",
    title: "Сотрудники",
    description: "Управление пользователями организации",
    icon: UsersIcon,
    path: "/organization/settings/users",
    color: "text-secondary",
  },
  {
    key: "knowledge",
    title: "База знаний",
    description: "Управление документами организации",
    icon: BookOpenIcon,
    path: "/organization/settings/knowledge",
    color: "text-primary",
  },
  {
    key: "memory",
    title: "Память агента",
    description: "Факты и контекст для AI ассистента",
    icon: LightBulbIcon,
    path: "/organization/settings/memory",
    color: "text-warning",
  },
  {
    key: "contracts",
    title: "Документы",
    description: "Сгенерированные договоры и контракты",
    icon: DocumentTextIcon,
    path: "/organization/settings/contracts",
    color: "text-success",
  },
  {
    key: "general",
    title: "Настройки",
    description: "Основные параметры организации",
    icon: Cog6ToothIcon,
    path: "/organization/settings/general",
    color: "text-default-400",
  },
];

export default function OrganizationPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations });
  const { core } = useApiClients();

  const [organization, setOrganization] = useState<CoreOrganization | null>(null);
  const [loadingOrg, setLoadingOrg] = useState(true);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, authLoading, router]);

  useEffect(() => {
    if (!authLoading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [authLoading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

  const loadOrganization = useCallback(async () => {
    if (!currentOrg?.id) return;

    setLoadingOrg(true);
    try {
      const response = await core.v1.organizationServiceGetOrganization(currentOrg.id);
      setOrganization(response.data.organization || null);
    } catch (e) {
      console.error("Failed to load organization", e);
    } finally {
      setLoadingOrg(false);
    }
  }, [core.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadOrganization();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadOrganization]);

  if (authLoading || orgLoading || loadingOrg) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner size="lg" />
      </div>
    );
  }

  if (!currentOrg || !organization) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-sm">
          <CardBody>
            <p className="text-default-400 text-center">Организация не найдена</p>
          </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen pb-20 gap-4">
      {/* Заголовок организации */}
      <Card className="rounded-4xl shadow-md mt-4 p-5 gap-3">
          <div className="flex items-start gap-2">
            <div className="flex-1 min-w-0 gap-0.5 flex flex-col">
              <h1 className="text-xl font-bold truncate">{organization.name || currentOrg.id}</h1>
              {(organization.industry || organization.region) && (
                <p className="text-xs text-default-400">
                  {organization.region}
                  {organization.industry && organization.region && " • "}
                  {organization.industry}
                </p>
              )}
            </div>
          </div>
          {organization.description && (
            <p className="text-xs text-default-400 font-semibold line-clamp-2">{organization.description}</p>
          )}
      </Card>

      {/* Список разделов */}
      <div className="flex-1">
        <h2 className="text-sm font-semibold text-default-400 uppercase tracking-wide mb-3">
          Разделы
        </h2>
        <div className="flex flex-col gap-3">
          {sections.map((section) => {
            const Icon = section.icon;

            return (
              <Card
                key={section.key}
                isPressable
                className="active:scale-[0.98] transition-transform rounded-4xl"
                shadow="sm"
                onPress={() => router.push(section.path)}
              >
                <CardBody className="p-4">
                  <div className="flex items-center gap-3">
                    <div className={`rounded-lg ${section.color}`}>
                      <Icon className="h-7 w-7" />
                    </div>
                    <div className="flex flex-col flex-1 min-w-0 gap-1">
                      <p className="font-semibold text-base">{section.title}</p>
                      <p className="text-xs text-default-400">{section.description}</p>
                    </div>
                    <ArrowRightIcon className="h-5 w-5 text-default-400 flex-shrink-0" />
                  </div>
                </CardBody>
              </Card>
            );
          })}
        </div>
      </div>
    </div>
  );
}
