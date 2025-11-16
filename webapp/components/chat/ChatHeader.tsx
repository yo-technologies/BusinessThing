"use client";

import { Card, CardHeader } from "@heroui/card";
import { Button } from "@heroui/button";
import { Bars3Icon, PlusIcon } from "@heroicons/react/24/outline";
import { AgentGetLLMLimitsResponse } from "@/api/api.agent.generated";

interface ChatHeaderProps {
  chatName?: string | null;
  limits?: AgentGetLLMLimitsResponse | null;
  usageTokens?: number | null;
  onShowChatList?: () => void;
  onCreateChat?: () => void;
}

export function ChatHeader({
  chatName,
  limits,
  usageTokens,
  onShowChatList,
  onCreateChat,
}: ChatHeaderProps) {
  // Вычисляем процент использования токенов
  const usagePercent = limits && limits.dailyLimit
    ? Math.min((((limits.used ?? 0) + (usageTokens ?? 0)) / limits.dailyLimit) * 100, 100)
    : 0;

  return (
    <Card className="relative border-none shadow-none overflow-visible rounded-b-none px-1">
      <CardHeader className="flex items-center gap-2 pb-1">
        <div className="flex flex-1 flex-col items-start gap-0.5">
          <span className="text-tiny font-medium uppercase text-default-400">Чат с агентом</span>
          <h1 className="text-base font-semibold">
            {chatName || "Новый чат"}
          </h1>
        </div>
        <div className="flex items-center gap-2">
          {onCreateChat && (
            <Button isIconOnly size="sm" variant="ghost" color="success" onPress={onCreateChat}>
              <PlusIcon className="h-5 w-5" />
            </Button>
          )}
          {onShowChatList && (
            <Button isIconOnly size="sm" variant="ghost" color="secondary" onPress={onShowChatList}>
              <Bars3Icon className="h-5 w-5" />
            </Button>
          )}
        </div>
      </CardHeader>
      
      {/* Индикатор использования токенов */}
      {limits && limits.dailyLimit && (
          <div className="mt-1 h-1 bg-default-200 rounded-full overflow-hidden">
            <div 
              className="h-full bg-primary transition-all duration-300"
              style={{ width: `${usagePercent}%` }}
            />
        </div>
      )}
    </Card>
  );
}
