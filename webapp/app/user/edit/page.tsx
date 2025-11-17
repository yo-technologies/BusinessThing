"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import { Input } from "@heroui/input";
import { Spinner } from "@heroui/spinner";
import { Card, CardBody, CardHeader, CardFooter } from "@heroui/card";

import { useAuth } from "@/hooks/useAuth";
import { useApiClients } from "@/api/client";

export default function EditUserPage() {
  const router = useRouter();
  const { userInfo, loading: authLoading } = useAuth();
  const { core } = useApiClients();

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      if (userInfo?.userId) {
        try {
          const response = await core.v1.userServiceGetUser(userInfo.userId);
          if (response.data.user) {
            setFirstName(response.data.user.firstName || "");
            setLastName(response.data.user.lastName || "");
          }
        } catch (err) {
          console.error("Failed to fetch user for editing", err);
          setError("Не удалось загрузить данные пользователя.");
        } finally {
          setIsLoading(false);
        }
      }
    };

    if (!authLoading) {
      fetchUser();
    }
  }, [userInfo, authLoading, core.v1]);

  const handleSave = async () => {
    if (!userInfo?.userId) {
      setError("Ошибка: ID пользователя не найден.");
      return;
    }

    setIsSaving(true);
    setError(null);

    try {
      await core.v1.authServiceCompleteRegistration({
        userId: userInfo.userId,
        firstName,
        lastName,
      });
      // Navigate back to the user page to see the changes
      router.push("/user");
    } catch (err) {
      console.error("Failed to update user", err);
      setError("Не удалось сохранить изменения. Попробуйте снова.");
      setIsSaving(false);
    }
  };

  if (isLoading || authLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Загрузка профиля..." />
      </div>
    );
  }

  return (
    <div className="flex justify-center items-start pt-12 px-4 h-full">
      <Card className="w-full max-w-md">
        <CardHeader>
          <h1 className="text-xl font-bold">Редактировать профиль</h1>
        </CardHeader>
        <CardBody className="flex flex-col gap-4">
          {error && (
            <div className="bg-danger-100 border border-danger-400 text-danger-700 px-4 py-3 rounded-lg">
              {error}
            </div>
          )}
          <Input
            label="Имя"
            value={firstName}
            onValueChange={setFirstName}
            placeholder="Введите ваше имя"
            isDisabled={isSaving}
          />
          <Input
            label="Фамилия"
            value={lastName}
            onValueChange={setLastName}
            placeholder="Введите вашу фамилию"
            isDisabled={isSaving}
          />
        </CardBody>
        <CardFooter className="flex justify-end gap-2">
          <Button variant="flat" onClick={() => router.back()} disabled={isSaving}>
            Отмена
          </Button>
          <Button color="primary" onClick={handleSave} disabled={isSaving}>
            {isSaving ? <Spinner color="white" size="sm" /> : "Сохранить"}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}