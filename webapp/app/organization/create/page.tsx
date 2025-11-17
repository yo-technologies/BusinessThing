"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import { Card } from "@heroui/card";
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

      const response =
        await core.v1.organizationServiceCreateOrganization(payload);

      if (response.data.organization?.id && typeof window !== "undefined") {
        localStorage.setItem(
          "businessthing_current_org_id",
          response.data.organization.id,
        );
      }

      await refreshAuthToken();
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
        <div className="space-y-4 overflow-auto p-4">
          <div className="flex flex-col items-start gap-1">
            <span className="text-tiny font-medium uppercase text-secondary">
              Первый шаг
            </span>
            <h1 className="text-xl font-semibold">Создай организацию</h1>
          </div>

          <p className="text-small text-default-400">
            Для работы с системой нужна организация. Укажи название компании и
            основную информацию.
          </p>

          {error && <p className="text-xs text-danger-500">{error}</p>}

          <Input
            autoFocus
            isRequired
            label="Название организации"
            placeholder="ООО «Бизнес»"
            radius="lg"
            value={name}
            variant="bordered"
            onChange={(e) => setName(e.target.value)}
          />

          <Input
            label="Отрасль"
            placeholder="IT, Производство, Торговля..."
            radius="lg"
            value={industry}
            variant="bordered"
            onChange={(e) => setIndustry(e.target.value)}
          />

          <Input
            label="Регион"
            placeholder="Москва, Санкт-Петербург..."
            radius="lg"
            value={region}
            variant="bordered"
            onChange={(e) => setRegion(e.target.value)}
          />

          <Textarea
            label="Описание"
            minRows={3}
            placeholder="Краткое описание деятельности компании"
            radius="lg"
            value={description}
            variant="bordered"
            onChange={(e) => setDescription(e.target.value)}
          />

          <Button
            className="mt-2 w-full"
            color="success"
            size="lg"
            isDisabled={!name.trim()}
            isLoading={submitting}
            radius="lg"
            onPress={handleSubmit}
          >
            Создать организацию
          </Button>
        </div>
      </Card>
    </div>
  );
}