"use client";

import { Card, CardBody, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Bars3Icon } from "@heroicons/react/24/outline";
import { AgentGetLLMLimitsResponse } from "@/api/api.agent.generated";

interface ChatHeaderProps {
  userName?: string;
  limits?: AgentGetLLMLimitsResponse | null;
  error?: string | null;
  usageTokens?: number | null;
  onShowChatList?: () => void;
}

export function ChatHeader({
  userName,
  limits,
  error,
  usageTokens,
  onShowChatList,
}: ChatHeaderProps) {
  return (
    <Card className="border-none bg-default-50/60 shadow-sm">
      <CardHeader className="flex items-center gap-2 pb-2">
        <div className="flex flex-1 flex-col items-start gap-1">
          <span className="text-tiny font-medium uppercase text-primary">Чат с агентом</span>
          <h1 className="text-base font-semibold">{userName ?? "Клиент"}</h1>
        </div>
        {limits && (
          <div className="flex flex-col items-end text-right text-[11px] text-default-500">
            <span>
              {limits.used ?? 0} / {limits.dailyLimit ?? 0}
            </span>
          </div>
        )}
        {onShowChatList && (
          <Button isIconOnly size="sm" variant="flat" onPress={onShowChatList}>
            <Bars3Icon className="h-5 w-5" />
          </Button>
        )}
      </CardHeader>
      {error && (
        <CardBody className="pb-3 pt-0">
          <p className="text-xs text-danger-500">{error}</p>
        </CardBody>
      )}
      {usageTokens && (
        <CardBody className="pb-3 pt-0">
          <p className="text-[11px] text-default-400">Потрачено токенов: {usageTokens}</p>
        </CardBody>
      )}
    </Card>
  );
}
