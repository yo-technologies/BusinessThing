"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Input } from "@heroui/input";

import { CoreCompleteRegistrationRequest } from "@/api/api.core.generated";
import { refreshAuthToken, useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";

export default function UserEditPage() {
  const router = useRouter();
  const { user, loading } = useAuth();
  const { core } = useApiClients();

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!loading && user) {
      setFirstName(user.firstName ?? "");
      setLastName(user.lastName ?? "");
    }
  }, [loading, user]);

  const handleUpdate = async () => {
    if (!firstName.trim() || !lastName.trim() || !user?.id) return;

    setSubmitting(true);
    try {
      const payload: CoreCompleteRegistrationRequest = {
        userId: user.id,
        firstName: firstName.trim(),
        lastName: lastName.trim(),
      };

      await core.v1.authServiceCompleteRegistration(payload);

      // Refresh the token to get updated user info in the JWT
      await refreshAuthToken();

      // Redirect back to the user profile page
      router.push("/user");
    } catch (error) {
      console.error("Failed to update user information:", error);
      // Optionally, show an error message to the user
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p>Loading user data...</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col items-center justify-center px-2">
      <Card className="w-full max-w-md border-none bg-content1/80 shadow-md">
        <CardHeader className="flex flex-col items-start gap-1 pb-2">
          <h1 className="text-xl font-semibold">Редактировать профиль</h1>
        </CardHeader>
        <CardBody className="space-y-4">
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
            color="primary"
            radius="lg"
            className="w-full mt-2"
            isDisabled={!firstName.trim() || !lastName.trim() || submitting}
            isLoading={submitting}
            onPress={handleUpdate}
          >
            Сохранить
          </Button>
        </CardBody>
      </Card>
    </div>
  );
}
