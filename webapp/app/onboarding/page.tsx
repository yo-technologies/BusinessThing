"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Input } from "@heroui/input";

import { CoreCompleteRegistrationRequest } from "@/api/api.core.generated";
import { refreshAuthToken, useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";

export default function OnboardingPage() {
  const router = useRouter();
  const { user, isAuthenticated, loading, isNewUser } = useAuth();
  const { core } = useApiClients();

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    // Если не новый пользователь или нет пользователя, редиректим
    if (!loading && (!isNewUser || !user?.id)) {
      router.replace("/chat");
    }
  }, [isNewUser, loading, router, user?.id]);

  useEffect(() => {
    if (user) {
      setFirstName(user.firstName ?? "");
      setLastName(user.lastName ?? "");
    }
  }, [user]);

  const handleSubmit = async () => {
    if (!firstName.trim() || !lastName.trim() || !user?.id) return;

    setSubmitting(true);
    try {
      const payload: CoreCompleteRegistrationRequest = {
        userId: user.id,
        firstName: firstName.trim(),
        lastName: lastName.trim(),
      };

      await core.v1.authServiceCompleteRegistration(payload);

      // Синхронно обновляем токен с актуальным списком организаций
      await refreshAuthToken();

      // Переходим на главную страницу (токен уже обновлен)
      router.push("/chat");
    } catch (error) {
      console.error("Failed to complete registration:", error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex h-full flex-col items-center justify-center px-2">
      <Card className="w-full max-w-md border-none bg-content1/80 shadow-md">
        <CardHeader className="flex flex-col items-start gap-1 pb-2">
          <span className="text-tiny font-medium uppercase text-secondary">Добро пожаловать</span>
          <h1 className="text-xl font-semibold">Заполни профиль</h1>
        </CardHeader>
        <CardBody className="space-y-4">
          <p className="text-small text-default-400">
            Мы сохраним твои имя и фамилию, чтобы подставлять их в документы и чаты.
          </p>

          <Input
            label="Имя"
            placeholder="Иван"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            variant="bordered"
            radius="lg"
            isRequired
          />

          <Input
            label="Фамилия"
            placeholder="Иванов"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            variant="bordered"
            radius="lg"
            isRequired
          />

          <Button
            color="success"
            radius="lg"
            className="w-full mt-2"
            isDisabled={!firstName.trim() || !lastName.trim()}
            isLoading={submitting}
            onPress={handleSubmit}
          >
            Продолжить
          </Button>
        </CardBody>
      </Card>
    </div>
  );
}
