"use client";

import { useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody } from "@heroui/card";
import { PlusIcon, ChatBubbleLeftIcon, TrashIcon } from "@heroicons/react/24/outline";
import { AgentChat } from "@/api/api.agent.generated";
import { Modal, ModalContent, ModalHeader, ModalBody, ModalFooter } from "@heroui/modal";
import { Spinner } from "@heroui/spinner";

interface ChatListProps {
  chats: AgentChat[];
  selectedChatId: string | null;
  onSelectChat: (chatId: string) => void;
  onCreateChat: () => void;
  onDeleteChat: (chatId: string) => Promise<void>;
  loading?: boolean;
}

export function ChatList({
  chats,
  selectedChatId,
  onSelectChat,
  onCreateChat,
  onDeleteChat,
  loading = false,
}: ChatListProps) {
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [chatToDelete, setChatToDelete] = useState<AgentChat | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDeleteClick = (chat: AgentChat, e: React.MouseEvent) => {
    e.stopPropagation();
    setChatToDelete(chat);
    setDeleteModalOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (!chatToDelete?.id) return;

    setIsDeleting(true);
    try {
      await onDeleteChat(chatToDelete.id);
      setDeleteModalOpen(false);
      setChatToDelete(null);
    } catch (error) {
      console.error("Failed to delete chat", error);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleCancelDelete = () => {
    setDeleteModalOpen(false);
    setChatToDelete(null);
  };

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-sm text-default-400">Загрузка чатов...</p>
      </div>
    );
  }

  return (
    <>
      <div className="flex h-full flex-col gap-2">
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
                isHoverable
                className={`border-none transition-all cursor-pointer ${
                  selectedChatId === chat.id
                    ? "bg-secondary/20 shadow-md"
                    : "bg-default-50/50"
                }`}
              >
                <CardBody className="py-3">
                  <div className="flex items-center justify-between gap-2">
                    <div 
                      className="flex-1 overflow-hidden cursor-pointer"
                      onClick={() => chat.id && onSelectChat(chat.id)}
                    >
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
                    <Button
                      isIconOnly
                      size="sm"
                      variant="light"
                      color="danger"
                      onClick={(e) => handleDeleteClick(chat, e)}
                      className="min-w-8 h-8"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                </CardBody>
              </Card>
            ))
          )}
        </div>
      </div>

      <Modal
        isOpen={deleteModalOpen}
        onClose={handleCancelDelete}
        backdrop="opaque"
      >
        <ModalContent>
          <ModalHeader className="flex flex-col gap-1">
            Удалить чат?
          </ModalHeader>
          <ModalBody>
            <p className="text-sm text-default-600">
              Вы уверены, что хотите удалить чат &quot;{chatToDelete?.title || "Новый чат"}&quot;?
              Это действие нельзя отменить.
            </p>
          </ModalBody>
          <ModalFooter>
            <Button
              variant="light"
              onPress={handleCancelDelete}
              isDisabled={isDeleting}
            >
              Отмена
            </Button>
            <Button
              color="danger"
              onPress={handleConfirmDelete}
              isLoading={isDeleting}
              spinner={<Spinner classNames={{wrapper:"w-3 h-3"}} size="sm" color="white" />}
            >
              Удалить
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  );
}
