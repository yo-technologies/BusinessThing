"use client";

import { Drawer, DrawerContent, DrawerHeader, DrawerBody } from "@heroui/drawer";
import { ChatList } from "./ChatList";
import { AgentChat } from "@/api/api.agent.generated";

interface ChatListModalProps {
  isOpen: boolean;
  onClose: () => void;
  chats: AgentChat[];
  selectedChatId: string | null;
  onSelectChat: (chatId: string) => void;
  onCreateChat: () => void;
  loading?: boolean;
}

export function ChatListModal({
  isOpen,
  onClose,
  chats,
  selectedChatId,
  onSelectChat,
  onCreateChat,
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
      isOpen={isOpen}
      onClose={onClose}
      placement="bottom"
      backdrop="blur"
      size="4xl"
      classNames={{
        base: "pb-safe",
        body: "px-4 pb-6 overflow-y-auto",
      }}
    >
      <DrawerContent>
        <DrawerHeader className="flex flex-col gap-1 px-4 pt-safe">
          <h2 className="text-lg font-semibold">Ваши чаты</h2>
        </DrawerHeader>
        <DrawerBody>
          <ChatList
            chats={chats}
            selectedChatId={selectedChatId}
            onSelectChat={handleSelectChat}
            onCreateChat={handleCreateChat}
            loading={loading}
          />
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  );
}
