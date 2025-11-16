"use client";

import { useAuth } from "@/hooks/useAuth";
import { Avatar } from "@heroui/avatar";
import { Button } from "@heroui/button";
import {
  PencilSquareIcon,
  BuildingOfficeIcon,
  ChevronRightIcon,
} from "@heroicons/react/24/outline";
import Link from "next/link";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { initData } from "@telegram-apps/sdk";
import { useEffect } from "react"; // Import useEffect

export default function UserPage() {
  const { user, loading } = useAuth();
  
  useEffect(() => { // Wrap initData.restore() in useEffect
    initData.restore();
  }, []); // Empty dependency array to run only once

  const telegramUser = initData.user()

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Загрузка..." color="primary" />
      </div>
    );
  }

  const photoUrl = telegramUser?.photo_url;

  return (
    <div className="flex flex-col items-center pt-12 px-4">
      <Avatar
        src={photoUrl || ""}
        className="w-24 h-24 text-large"
      />
      <h1 className="text-2xl font-bold mt-4">{`${user?.firstName || "Имя"} ${
        user?.lastName || "Фамилия"
      }`}</h1>
      <Divider className="mt-2"/>
      <div className="w-full mt-8 space-y-2">
        <Button
          as={Link}
          href="/user/edit"
          className="w-full flex justify-between items-center bg-zinc-800/70 text-white py-6"
        >
          <div className="flex items-center gap-3">
            <PencilSquareIcon className="w-6 h-6" />
            <span>Редактировать профиль</span>
          </div>
          <ChevronRightIcon className="w-5 h-5" />
        </Button>
        <Button
          as={Link}
          href="/organization"
          className="w-full flex justify-between items-center bg-zinc-800/70 text-white py-6"
        >
          <div className="flex items-center gap-3">
            <BuildingOfficeIcon className="w-6 h-6" />
            <span>Мои организации</span>
          </div>
          <ChevronRightIcon className="w-5 h-5" />
        </Button>
      </div>
    </div>
  );
}