"use client";

import { useCallback, useEffect, useState } from "react";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useChatList } from "@/hooks/useChatList";
import { useChatWebSocket } from "@/hooks/useChatWebSocket";
import { useApiClients } from "@/api/client";
import { ChatWindow, ChatInput, ChatHeader, ChatListModal } from "@/components/chat";
import { AgentGetLLMLimitsResponse } from "@/api/api.agent.generated";

export default function ChatPage() {
  const router = useRouter();
  const { isAuthenticated, loading, user, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations });
  const { agent } = useApiClients();

  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [showChatListModal, setShowChatListModal] = useState(false);
  const [input, setInput] = useState("");
  const [limits, setLimits] = useState<AgentGetLLMLimitsResponse | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);

  const { chats, loading: chatsLoading, reload: reloadChats } = useChatList({
    organizationId: currentOrg?.id,
    enabled: isAuthenticated && !isNewUser && !needsOrganization,
  });

  const {
    messages,
    streamingMessage,
    isStreaming,
    error,
    usageTokens,
    sendMessage,
    setMessages,
    clearError,
  } = useChatWebSocket({
    chatId: selectedChatId,
    organizationId: currentOrg?.id,
    onChatCreated: useCallback((newChatId: string) => {
      setSelectedChatId(newChatId);
      reloadChats();
    }, [reloadChats]),
  });

  useEffect(() => {
    if (!loading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, loading, router]);

  useEffect(() => {
    if (!loading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [loading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

  useEffect(() => {
    const loadInitialData = async () => {
      if (!currentOrg?.id || !isAuthenticated || isNewUser) return;

      try {
        const limitsResp = await agent.v1.agentServiceGetLlmLimits();
        setLimits(limitsResp.data ?? null);

        // По умолчанию показываем новый чат (selectedChatId = null)
        // Пользователь может открыть список чатов по кнопке
        setSelectedChatId(null);
      } catch (e) {
        console.error("Failed to load initial data", e);
      } finally {
        setInitialLoading(false);
      }
    };

    if (!chatsLoading && chats.length >= 0) {
      void loadInitialData();
    }
  }, [agent.v1, currentOrg?.id, isAuthenticated, isNewUser, chats, chatsLoading]);

  const handleSelectChat = useCallback(
    async (chatId: string) => {
      setSelectedChatId(chatId);
      clearError();

      try {
        const messagesResp = await agent.v1.agentServiceGetMessages(chatId, { limit: 50, offset: 0 });
        setMessages(messagesResp.data.messages ?? []);
      } catch (e) {
        console.error("Failed to load messages", e);
      }
    },
    [agent.v1, setMessages, clearError]
  );

  const handleCreateChat = useCallback(() => {
    setSelectedChatId(null);
    setMessages([]);
    clearError();
  }, [setMessages, clearError]);

  const handleSend = useCallback(() => {
    if (!input.trim()) return;

    sendMessage(input.trim());
    setInput("");
  }, [input, sendMessage]);

  if (loading || orgLoading || initialLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Загружаем чат..." color="primary" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3">
        <h1 className="text-xl font-semibold">Не удалось авторизоваться</h1>
        <p className="text-center text-small text-default-500">
          Попробуй закрыть мини-приложение и открыть его заново из Telegram.
        </p>
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-3">
        {/* <ChatHeader
          userName={user?.firstName}
          limits={limits}
          error={error}
          usageTokens={usageTokens}
          onShowChatList={() => setShowChatListModal(true)}
        /> */}

        <ChatWindow messages={messages} streamingMessage={streamingMessage} isStreaming={isStreaming} />

        <ChatInput
          value={input}
          onChange={setInput}
          onSend={handleSend}
          disabled={false}
          isStreaming={isStreaming}
        />
      </div>

      <ChatListModal
        isOpen={showChatListModal}
        onClose={() => setShowChatListModal(false)}
        chats={chats}
        selectedChatId={selectedChatId}
        onSelectChat={handleSelectChat}
        onCreateChat={handleCreateChat}
        loading={chatsLoading}
      />
    </>
  );
}
