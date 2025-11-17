"use client";

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerBody,
} from "@heroui/drawer";

import { ChatList } from "./ChatList";

import { AgentChat } from "@/api/api.agent.generated";

interface ChatListModalProps {
  isOpen: boolean;
  onClose: () => void;
  chats: AgentChat[];
  selectedChatId: string | null;
  onSelectChat: (chatId: string) => void;
  onCreateChat: () => void;
  onDeleteChat: (chatId: string) => Promise<void>;
  loading?: boolean;
}

export function ChatListModal({
  isOpen,
  onClose,
  chats,
  selectedChatId,
  onSelectChat,
  onCreateChat,
  onDeleteChat,
  loading,
}: ChatListModalProps) {
  const handleSelectChat = (chatId: string) => {
    onSelectChat(chatId);
    onClose();
  };

  const handleCreateChat = () => {
    onCreateChat();
    onClose();
  };

  return (
    <Drawer
      backdrop="blur"
      classNames={{
        base: "pb-safe",
        body: "px-4 pb-6 overflow-y-auto",
      }}
      isOpen={isOpen}
      placement="bottom"
      size="2xl"
      onClose={onClose}
    >
      <DrawerContent>
        <DrawerHeader className="flex flex-col gap-1 px-4 pt-4 pb-2 border-b border-default-200">
          <h2 className="text-lg font-semibold">Ваши чаты</h2>
        </DrawerHeader>
        <DrawerBody>
          <ChatList
            chats={chats}
            loading={loading}
            selectedChatId={selectedChatId}
            onCreateChat={handleCreateChat}
            onDeleteChat={onDeleteChat}
            onSelectChat={handleSelectChat}
          />
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  );
}
