"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Input, Textarea } from "@heroui/input";
import { Cog6ToothIcon } from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { CoreOrganization } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";
import { useCurrentRole } from "@/hooks/useCurrentRole";

export default function GeneralSettingsPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations, authLoading });
  const { core } = useApiClients();
  const { isAdmin } = useCurrentRole();

  const [organization, setOrganization] = useState<CoreOrganization | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [name, setName] = useState("");
  const [industry, setIndustry] = useState("");
  const [region, setRegion] = useState("");
  const [description, setDescription] = useState("");

  useBackButton(true);

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

    setInitialLoading(true);
    setError(null);
    try {
      const response = await core.v1.organizationServiceGetOrganization(currentOrg.id);
      const org = response.data.organization;
      if (org) {
        setOrganization(org);
        setName(org.name || "");
        setIndustry(org.industry || "");
        setRegion(org.region || "");
        setDescription(org.description || "");
      }
    } catch (e) {
      console.error("Failed to load organization", e);
      setError("Не удалось загрузить данные организации");
    } finally {
      setInitialLoading(false);
    }
  }, [core.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadOrganization();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadOrganization]);

  const handleSave = async () => {
    if (!organization?.id || !isAdmin) return;

    setSaving(true);
    try {
      const response = await core.v1.organizationServiceUpdateOrganization(organization.id, {
        name: name || undefined,
        industry: industry || undefined,
        region: region || undefined,
        description: description || undefined,
      });
      if (response.data.organization) {
        setOrganization(response.data.organization);
      }
    } catch (e) {
      console.error("Failed to update organization", e);
      setError("Не удалось сохранить изменения");
    } finally {
      setSaving(false);
    }
  };

  const hasChanges = organization && (
    name !== (organization.name || "") ||
    industry !== (organization.industry || "") ||
    region !== (organization.region || "") ||
    description !== (organization.description || "")
  );

  if (authLoading || orgLoading || initialLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <Card className="max-w-xl mx-auto mt-8 rounded-4xl shadow-sm border border-danger-200/60">
        <CardBody>
          <p className="text-danger text-center">{error}</p>
        </CardBody>
      </Card>
    );
  }

  if (!organization) {
    return (
      <Card className="max-w-xl mx-auto mt-8 rounded-4xl shadow-sm border border-default-100/60">
        <CardBody>
          <p className="text-default-400 text-center">Организация не найдена</p>
        </CardBody>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4 flex-1">
      <Card className="rounded-4xl shadow-none">
        <CardHeader className="flex flex-col gap-1 px-5 py-4">
          <div className="flex items-start gap-2 w-full">
            <Cog6ToothIcon className="h-5 w-5 flex-shrink-0 text-default mt-1" />
            <p className="text-xl font-semibold">Настройки</p>
          </div>
          <p className="text-xs text-default-300 text-left w-full">
            Основная информация об организации
          </p>
        </CardHeader>
      </Card>

      <Card className="flex-1 rounded-4xl shadow-none ">
        <CardBody className="gap-5 px-5 py-5">
          <Input
            label="Название организации"
            placeholder="ООО «Моя компания»"
            value={name}
            onValueChange={setName}
            isDisabled={!isAdmin}
          />

          <Input
            label="Отрасль"
            placeholder="IT, Финансы, Производство..."
            value={industry}
            onValueChange={setIndustry}
            isDisabled={!isAdmin}
          />

          <Input
            label="Регион"
            placeholder="Москва, Санкт-Петербург..."
            value={region}
            onValueChange={setRegion}
            isDisabled={!isAdmin}
          />

          <Textarea
            label="Описание"
            placeholder="Краткое описание деятельности организации"
            value={description}
            onValueChange={setDescription}
            minRows={4}
            maxRows={10}
            isDisabled={!isAdmin}
          />

          {!isAdmin && (
            <p className="text-sm text-default-400">
              Только администраторы могут редактировать настройки организации
            </p>
          )}

          <div className="rounded-3xl  bg-default-50/40 px-4 py-4">
            <h3 className="text-base font-semibold mb-3">Дополнительная информация</h3>
            <div className="grid gap-3 text-sm">
              <div className="flex justify-between">
                <span className="text-default-400">ID организации:</span>
                <span className="font-mono">{organization.id}</span>
              </div>
              {organization.createdAt && (
                <div className="flex justify-between">
                  <span className="text-default-400">Дата создания:</span>
                  <span>{new Date(organization.createdAt).toLocaleString("ru-RU")}</span>
                </div>
              )}
              {organization.updatedAt && (
                <div className="flex justify-between">
                  <span className="text-default-400">Последнее обновление:</span>
                  <span>{new Date(organization.updatedAt).toLocaleString("ru-RU")}</span>
                </div>
              )}
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  );
}
