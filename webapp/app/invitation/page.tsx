"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";

import { refreshAuthToken, useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";

function InvitationContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token");
  
  const { isAuthenticated, loading: authLoading, isNewUser } = useAuth();
  const { core } = useApiClients();

  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [authLoading, isNewUser, router]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      setError("Необходимо авторизоваться для принятия приглашения");
    }
  }, [authLoading, isAuthenticated]);

  const handleAccept = async () => {
    if (!token || !isAuthenticated) return;

    setAccepting(true);
    setError(null);

    try {
      await core.v1.userServiceAcceptInvitation(token, {});

      // Синхронно обновляем токен с новыми организациями
      await refreshAuthToken();

      // Переходим на главную страницу (токен уже обновлен)
      router.push("/chat");
    } catch (err: any) {
      console.error("Failed to accept invitation:", err);
      setError(err.response?.data?.message || "Не удалось принять приглашение");
    } finally {
      setAccepting(false);
    }
  };

  if (authLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner size="lg" />
      </div>
    );
  }

  if (!token) {
    return (
      <div className="flex h-full flex-col items-center justify-center px-2">
        <Card className="w-full max-w-md border-none bg-content1/80 shadow-md">
          <CardBody className="text-center">
            <p className="text-default-400">Токен приглашения не найден</p>
          </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col items-center justify-center px-2">
      <Card className="w-full max-w-md border-none bg-content1/80 shadow-md">
        <CardHeader className="flex flex-col items-start gap-1 pb-2">
          <span className="text-tiny font-medium uppercase text-secondary">Приглашение</span>
          <h1 className="text-xl font-semibold">Присоединиться к организации</h1>
        </CardHeader>
        <CardBody className="space-y-4">
          {error ? (
            <div className="rounded-lg bg-danger-50 p-3 text-small text-danger">
              {error}
            </div>
          ) : (
            <p className="text-small text-default-400">
              Вы получили приглашение присоединиться к организации. Нажмите кнопку ниже, чтобы принять приглашение.
            </p>
          )}

          <Button
            color="success"
            radius="lg"
            className="w-full"
            isLoading={accepting}
            isDisabled={!isAuthenticated || Boolean(error)}
            onPress={handleAccept}
          >
            {accepting ? "Принимаю..." : "Принять приглашение"}
          </Button>

          <Button
            variant="light"
            radius="lg"
            className="w-full"
            onPress={() => router.push("/chat")}
          >
            Отмена
          </Button>
        </CardBody>
      </Card>
    </div>
  );
}

export default function InvitationPage() {
  return (
    <Suspense
      fallback={
        <div className="flex h-full items-center justify-center">
          <Spinner size="lg" />
        </div>
      }
    >
      <InvitationContent />
    </Suspense>
  );
}
