"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { retrieveLaunchParams } from "@telegram-apps/sdk-react";

import { refreshAuthToken, useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { markInvitationAsProcessed } from "@/hooks/useInvitationToken";

function InvitationContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const urlToken = searchParams.get("token");

  // Получаем токен из URL или из startParam (при запуске через Telegram)
  const [token, setToken] = useState<string | null>(urlToken);

  const { isAuthenticated, loading: authLoading, isNewUser } = useAuth();
  const { core } = useApiClients();

  // Проверяем startParam при монтировании
  useEffect(() => {
    if (typeof window === "undefined") return;

    try {
      const params = retrieveLaunchParams();
      // startParam может быть в params.startParam или в params.tgWebAppData.start_param
      const startParam =
        params.startParam || (params as any).tgWebAppData?.start_param;

      if (startParam && !urlToken) {
        // startParam может содержать токен приглашения в формате invitation_<token>
        const startParamStr = String(startParam);

        if (startParamStr.startsWith("invitation_")) {
          const invitationToken = startParamStr.replace("invitation_", "");

          setToken(invitationToken);
        }
      }
    } catch (error) {
      console.error(
        "[InvitationPage] Failed to retrieve launch params:",
        error,
      );
    }
  }, [urlToken]);

  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [autoAcceptAttempted, setAutoAcceptAttempted] = useState(false);

  const handleAccept = async () => {
    if (!token || !isAuthenticated) return;

    setAccepting(true);
    setError(null);

    try {
      await core.v1.userServiceAcceptInvitation(token, {});

      // Помечаем приглашение как обработанное
      markInvitationAsProcessed();

      // Синхронно обновляем токен с новыми организациями
      await refreshAuthToken();

      // Переходим на главную страницу (токен уже обновлен)
      router.push("/chat");
    } catch (err: any) {
      console.error("[InvitationPage] Failed to accept invitation:", err);
      console.error("[InvitationPage] Error status:", err.response?.status);
      console.error("[InvitationPage] Error data:", err.response?.data);

      // Если 400 - возможно приглашение уже использовано, логируем и идем дальше
      if (err.response?.status === 400) {
        console.warn(
          "[InvitationPage] Got 400 error, possibly invitation already used. Continuing anyway...",
        );

        // Помечаем приглашение как обработанное
        markInvitationAsProcessed();

        // Обновляем токен на всякий случай
        try {
          await refreshAuthToken();
        } catch (refreshErr) {
          console.error(
            "[InvitationPage] Failed to refresh token:",
            refreshErr,
          );
        }

        // Переходим на главную страницу
        router.push("/chat");
      } else {
        setError(
          err.response?.data?.message || "Не удалось принять приглашение",
        );
      }
    } finally {
      setAccepting(false);
    }
  };

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

  // Автоматическое принятие приглашения при переходе через startParam
  useEffect(() => {
    console.log("[InvitationPage] Auto-accept check:", {
      authLoading,
      isAuthenticated,
      isNewUser,
      token,
      accepting,
      autoAcceptAttempted,
      error,
      urlToken,
    });

    if (
      !authLoading &&
      isAuthenticated &&
      !isNewUser &&
      token &&
      !accepting &&
      !autoAcceptAttempted &&
      !error
    ) {
      // Проверяем, что токен пришел из startParam (не из URL)
      if (!urlToken) {
        console.log(
          "[InvitationPage] Auto-accepting invitation from startParam",
        );
        setAutoAcceptAttempted(true);
        handleAccept();
      } else {
        console.log("[InvitationPage] Not auto-accepting, token from URL");
      }
    }
  }, [
    authLoading,
    isAuthenticated,
    isNewUser,
    token,
    accepting,
    autoAcceptAttempted,
    error,
    urlToken,
  ]);

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
          <span className="text-tiny font-medium uppercase text-secondary">
            Приглашение
          </span>
          <h1 className="text-xl font-semibold">
            Присоединиться к организации
          </h1>
        </CardHeader>
        <CardBody className="space-y-4">
          {error ? (
            <div className="rounded-lg bg-danger-50 p-3 text-small text-danger">
              {error}
            </div>
          ) : (
            <p className="text-small text-default-400">
              Вы получили приглашение присоединиться к организации. Нажмите
              кнопку ниже, чтобы принять приглашение.
            </p>
          )}

          <Button
            className="w-full"
            color="success"
            isDisabled={!isAuthenticated || Boolean(error)}
            isLoading={accepting}
            radius="lg"
            onPress={handleAccept}
          >
            {accepting ? "Принимаю..." : "Принять приглашение"}
          </Button>

          <Button
            className="w-full"
            radius="lg"
            variant="light"
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
