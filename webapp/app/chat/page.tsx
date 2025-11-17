"use client";

import { useCallback, useEffect, useState } from "react";
import { Spinner } from "@heroui/spinner";
import { useRouter } from "next/navigation";
import { Card } from "@heroui/card";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useHasInvitation } from "@/hooks/useInvitationToken";
import { useChatList } from "@/hooks/useChatList";
import { useChatWebSocket } from "@/hooks/useChatWebSocket";
import { useApiClients } from "@/api/client";
import {
  ChatWindow,
  ChatInput,
  ChatHeader,
  ChatListModal,
} from "@/components/chat";
import { AgentGetLLMLimitsResponse } from "@/api/api.agent.generated";

const CHAT_SESSION_KEY = "last_chat_id";

export default function ChatPage() {
  const router = useRouter();
  const { isAuthenticated, loading, user, isNewUser, organizations } =
    useAuth();
  const {
    currentOrg,
    loading: orgLoading,
    needsOrganization,
  } = useOrganization({ organizations, authLoading: loading });
  const hasInvitation = useHasInvitation();
  const { agent } = useApiClients();

  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [showChatListModal, setShowChatListModal] = useState(false);
  const [input, setInput] = useState("");
  const [limits, setLimits] = useState<AgentGetLLMLimitsResponse | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const {
    chats,
    loading: chatsLoading,
    reload: reloadChats,
    deleteChat,
  } = useChatList({
    organizationId: currentOrg?.id,
    enabled: isAuthenticated && !isNewUser && !needsOrganization,
  });

  const {
    messages,
    streamingMessage,
    streamingToolCalls,
    isStreaming,
    error,
    usageTokens,
    chatName,
    currentChatId,
    sendMessage,
    setMessages,
    setChatName,
    clearError,
  } = useChatWebSocket({
    chatId: selectedChatId,
    organizationId: currentOrg?.id,
    onChatCreated: useCallback(
      (newChatId: string) => {
        setSelectedChatId(newChatId);

        // Сохраняем новый chatId в sessionStorage
        sessionStorage.setItem(CHAT_SESSION_KEY, newChatId);

        reloadChats();
      },
      [reloadChats, selectedChatId],
    ),
    onFinalReceived: useCallback(async () => {
      // Перезагружаем лимиты после завершения стрима
      try {
        const limitsResp = await agent.v1.agentServiceGetLlmLimits();

        setLimits(limitsResp.data ?? null);
      } catch (e) {
        console.error("Failed to reload limits", e);
      }
    }, [agent.v1]),
  });

  useEffect(() => {
    if (!loading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, loading, router]);

  // Проверяем наличие приглашения перед редиректом на создание организации
  useEffect(() => {
    if (
      !loading &&
      !orgLoading &&
      isAuthenticated &&
      !isNewUser
    ) {
      // Приоритет - приглашение (даже если у пользователя уже есть организации)
      if (hasInvitation) {
        router.replace("/invitation");
      } else if (needsOrganization) {
        router.replace("/organization/create");
      }
    }
  }, [
    loading,
    orgLoading,
    isAuthenticated,
    isNewUser,
    needsOrganization,
    hasInvitation,
    router,
  ]);

  // Инициализация при первой загрузке
  useEffect(() => {
    const loadInitialData = async () => {
      if (
        !currentOrg?.id ||
        !isAuthenticated ||
        isNewUser ||
        chatsLoading ||
        needsOrganization
      )
        return;

      try {
        const limitsResp = await agent.v1.agentServiceGetLlmLimits();

        setLimits(limitsResp.data ?? null);
        setInitialLoading(false);
      } catch (e) {
        console.error("Failed to load initial data", e);
        setInitialLoading(false);
      }
    };

    // Выполняем только один раз при монтировании компонента
    if (initialLoading && !chatsLoading && !orgLoading) {
      void loadInitialData();
    }
  }, [
    agent.v1,
    currentOrg?.id,
    isAuthenticated,
    isNewUser,
    chatsLoading,
    needsOrganization,
    initialLoading,
    orgLoading,
  ]);

  const handleSelectChat = useCallback(
    async (chatId: string) => {
      setSelectedChatId(chatId);
      clearError();
      setLoadingMessages(true);

      // Сохраняем chatId в sessionStorage
      sessionStorage.setItem(CHAT_SESSION_KEY, chatId);

      try {
        // Загружаем сообщения
        const messagesResp = await agent.v1.agentServiceGetMessages(chatId, {
          orgId: currentOrg?.id ?? "",
          limit: 50,
          offset: 0,
        });

        // Переворачиваем массив, т.к. backend отдаёт от новых к старым
        setMessages([...(messagesResp.data.messages ?? [])].reverse());

        // Загружаем информацию о чате
        const chatResp = await agent.v1.agentServiceGetChat(chatId, {
          orgId: currentOrg?.id ?? "",
        });

        if (chatResp.data.chat?.title) {
          setChatName(chatResp.data.chat.title);
        }
      } catch (e) {
        console.error("Failed to load chat data", e);
      } finally {
        setLoadingMessages(false);
      }
    },
    [agent.v1, currentOrg?.id, setMessages, setChatName, clearError],
  );

  // Восстановление последнего чата из sessionStorage
  useEffect(() => {
    if (
      initialLoading ||
      chatsLoading ||
      orgLoading ||
      !chats.length ||
      !currentOrg?.id
    )
      return;

    const lastChatId = sessionStorage.getItem(CHAT_SESSION_KEY);

    if (lastChatId && !selectedChatId) {
      // Проверяем, что чат существует в списке
      const chatExists = chats.some((chat) => chat.id === lastChatId);

      if (chatExists) {
        void handleSelectChat(lastChatId);
      } else {
      }
    }
  }, [
    initialLoading,
    chatsLoading,
    orgLoading,
    chats,
    selectedChatId,
    currentOrg?.id,
    handleSelectChat,
  ]);

  const handleCreateChat = useCallback(() => {
    setSelectedChatId(null);
    setMessages([]);
    setChatName(null);
    clearError();

    // Очищаем sessionStorage при создании нового чата
    sessionStorage.removeItem(CHAT_SESSION_KEY);
  }, [setMessages, setChatName, clearError]);

  const handleDeleteChat = useCallback(
    async (chatId: string) => {
      await deleteChat(chatId);

      // Если удалили активный чат - очищаем состояние
      if (selectedChatId === chatId) {
        setSelectedChatId(null);
        setMessages([]);
        setChatName(null);
        sessionStorage.removeItem(CHAT_SESSION_KEY);
      }
    },
    [deleteChat, selectedChatId, setMessages, setChatName],
  );

  const handleSend = useCallback(() => {
    if (!input.trim()) return;

    sendMessage(input.trim());
    setInput("");
  }, [input, sendMessage]);

  if (loading || orgLoading || initialLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner color="primary" label="Загружаем чат..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3">
        <h1 className="text-xl font-semibold">Не удалось авторизоваться</h1>
        <p className="text-center text-small text-default-400">
          Попробуй закрыть мини-приложение и открыть его заново из Telegram.
        </p>
      </div>
    );
  }

  return (
    <div className="flex h-full px-4">
      <Card className="flex flex-1 flex-col rounded-4xl shadow-none py-0">
        <ChatHeader
          chatName={chatName}
          limits={limits}
          usageTokens={usageTokens}
          onCreateChat={handleCreateChat}
          onShowChatList={() => setShowChatListModal(true)}
        />

        <ChatWindow
          chatId={selectedChatId}
          isStreaming={isStreaming}
          loadingMessages={loadingMessages}
          messages={messages}
          streamingMessage={streamingMessage}
          streamingToolCalls={streamingToolCalls}
        />

        <ChatInput
          disabled={false}
          isStreaming={isStreaming}
          value={input}
          onChange={setInput}
          onSend={handleSend}
        />
      </Card>

      <ChatListModal
        chats={chats}
        isOpen={showChatListModal}
        loading={chatsLoading}
        selectedChatId={selectedChatId}
        onClose={() => setShowChatListModal(false)}
        onCreateChat={handleCreateChat}
        onDeleteChat={handleDeleteChat}
        onSelectChat={handleSelectChat}
      />
    </div>
  );
}
