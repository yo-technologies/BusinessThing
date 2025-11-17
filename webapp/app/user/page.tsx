"use client";

import { useAuth } from "@/hooks/useAuth";
import { Avatar } from "@heroui/avatar";
import { Card, CardBody } from "@heroui/card";
import {
  PencilSquareIcon,
  BuildingOfficeIcon,
  ArrowRightIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { initData, User } from "@telegram-apps/sdk";
import { useEffect, useState } from "react";
import { useApiClients } from "@/api/client";
import { CoreUser } from "@/api/api.core.generated";

const userSections = [
  {
    key: "edit-profile",
    title: "Редактировать профиль",
    description: "Изменить имя и фамилию",
    icon: PencilSquareIcon,
    path: "/user/edit",
    color: "text-primary",
  },
  {
    key: "my-organizations",
    title: "Мои организации",
    description: "Просмотр и управление",
    icon: BuildingOfficeIcon,
    path: "/user/organizations", // Changed path to /organizations
    color: "text-secondary",
  },
];

export default function UserPage() {
  const { userInfo, loading: authLoading } = useAuth();
  const router = useRouter();
  const { core } = useApiClients();
  const [telegramUser, setTelegramUser] = useState<User | undefined>();
  const [detailedUser, setDetailedUser] = useState<CoreUser | undefined>();

  useEffect(() => {
    const fetchTelegramUser = async () => {
      await initData.restore();
      setTelegramUser(initData.user());
    };

    void fetchTelegramUser();
  }, []);

  useEffect(() => {
    const fetchDetailedUser = async () => {
      if (userInfo?.userId) {
        try {
          const response = await core.v1.userServiceGetUser(userInfo.userId);
          if (response.data.user) {
            setDetailedUser(response.data.user);
          }
        } catch (error) {
          console.error("Failed to fetch detailed user", error);
        }
      }
    };

    void fetchDetailedUser();
  }, [userInfo, core.v1]);

  const isLoading = authLoading || !telegramUser || !detailedUser;

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Загрузка..." color="primary" />
      </div>
    );
  }

  const photoUrl = telegramUser?.photo_url;
  const displayName =
    detailedUser?.firstName || detailedUser?.lastName
      ? `${detailedUser.firstName || ""} ${detailedUser.lastName || ""}`.trim()
      : `${telegramUser?.first_name || ""} ${
          telegramUser?.last_name || ""
        }`.trim();

  return (
    <div className="flex flex-col items-center pt-12 px-4">
      <Avatar src={photoUrl || ""} className="w-24 h-24 text-large" />
      <h1 className="text-2xl font-bold mt-4">
        {displayName || "Профиль"}
      </h1>
      <Divider className="mt-2" />
      <div className="w-full max-w-md mt-8">
        <div className="flex flex-col gap-3">
          {userSections.map((section) => {
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
                      <p className="text-xs text-default-400">
                        {section.description}
                      </p>
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