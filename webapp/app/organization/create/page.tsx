"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Input, Textarea } from "@heroui/input";

import { CoreCreateOrganizationRequest } from "@/api/api.core.generated";
import { refreshAuthToken, useApiClients } from "@/api/client";

export default function CreateOrganizationPage() {
  const router = useRouter();
  const { core } = useApiClients();

  const [name, setName] = useState("");
  const [industry, setIndustry] = useState("");
  const [region, setRegion] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!name.trim()) return;

    setSubmitting(true);
    setError(null);

    try {
      const payload: CoreCreateOrganizationRequest = {
        name: name.trim(),
        industry: industry.trim() || undefined,
        region: region.trim() || undefined,
        description: description.trim() || undefined,
      };

      const response = await core.v1.organizationServiceCreateOrganization(payload);

      if (response.data.organization?.id && typeof window !== "undefined") {
        localStorage.setItem("businessthing_current_org_id", response.data.organization.id);
      }

      // Синхронно обновляем токен с актуальным списком организаций
      await refreshAuthToken();

      // Переходим на главную страницу (токен уже обновлен)
      router.push("/chat");
    } catch (e) {
      console.error("Failed to create organization", e);
      setError("Не удалось создать организацию");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex h-full flex-col items-center justify-center px-2">
      <Card className="w-full max-w-md border-none bg-content1/80 shadow-md">
        <CardHeader className="flex flex-col items-start gap-1 pb-2">
          <span className="text-tiny font-medium uppercase text-secondary">Первый шаг</span>
          <h1 className="text-xl font-semibold">Создай организацию</h1>
        </CardHeader>
        <CardBody className="space-y-4">
          <p className="text-small text-default-400">
            Для работы с системой нужна организация. Укажи название компании и основную информацию.
          </p>

          {error && <p className="text-xs text-danger-500">{error}</p>}

          <Input
            label="Название организации"
            placeholder="ООО «Бизнес»"
            value={name}
            onChange={(e) => setName(e.target.value)}
            variant="bordered"
            radius="lg"
            isRequired
            autoFocus
          />

          <Input
            label="Отрасль"
            placeholder="IT, Производство, Торговля..."
            value={industry}
            onChange={(e) => setIndustry(e.target.value)}
            variant="bordered"
            radius="lg"
          />

          <Input
            label="Регион"
            placeholder="Москва, Санкт-Петербург..."
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            variant="bordered"
            radius="lg"
          />

          <Textarea
            label="Описание"
            placeholder="Краткое описание деятельности компании"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            variant="bordered"
            radius="lg"
            minRows={3}
          />

          <Button
            color="success"
            radius="lg"
            className="mt-2 w-full"
            isDisabled={!name.trim()}
            isLoading={submitting}
            onPress={handleSubmit}
          >
            Создать организацию
          </Button>
        </CardBody>
      </Card>
    </div>
  );
}
