"use client";

import { Button } from "@heroui/button";
import { Card, CardBody } from "@heroui/card";
import { PlusIcon, ChatBubbleLeftIcon } from "@heroicons/react/24/outline";
import { AgentChat } from "@/api/api.agent.generated";

interface ChatListProps {
  chats: AgentChat[];
  selectedChatId: string | null;
  onSelectChat: (chatId: string) => void;
  onCreateChat: () => void;
  loading?: boolean;
}

export function ChatList({
  chats,
  selectedChatId,
  onSelectChat,
  onCreateChat,
  loading = false,
}: ChatListProps) {
  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-default-400">Загрузка чатов...</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-2">
      <div className="flex items-center justify-between px-1">
        <h2 className="text-lg font-semibold">Чаты</h2>
        <Button
          isIconOnly
          size="sm"
          variant="flat"
          color="primary"
          onPress={onCreateChat}
        >
          <PlusIcon className="h-5 w-5" />
        </Button>
      </div>

      <div className="flex flex-1 flex-col gap-2 overflow-y-auto">
        {chats.length === 0 ? (
          <Card className="border-none bg-default-50/50">
            <CardBody className="flex flex-col items-center justify-center gap-2 py-8">
              <ChatBubbleLeftIcon className="h-8 w-8 text-default-300" />
              <p className="text-center text-sm text-default-400">
                Нет чатов. Создайте новый!
              </p>
            </CardBody>
          </Card>
        ) : (
          chats.map((chat) => (
            <Card
              key={chat.id}
              isPressable
              isHoverable
              onPress={() => chat.id && onSelectChat(chat.id)}
              className={`border-none transition-all ${
                selectedChatId === chat.id
                  ? "bg-primary/20 shadow-md"
                  : "bg-default-50/50"
              }`}
            >
              <CardBody className="py-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex-1 overflow-hidden">
                    <p className="truncate text-sm font-medium">
                      {chat.title || "Новый чат"}
                    </p>
                    {chat.updatedAt && (
                      <p className="text-xs text-default-400">
                        {new Date(chat.updatedAt).toLocaleDateString("ru-RU", {
                          day: "numeric",
                          month: "short",
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </p>
                    )}
                  </div>
                </div>
              </CardBody>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
