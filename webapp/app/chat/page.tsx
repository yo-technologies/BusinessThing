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
import { Card } from "@heroui/card";

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
  const [loadingMessages, setLoadingMessages] = useState(false);

  const { chats, loading: chatsLoading, reload: reloadChats } = useChatList({
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
    onChatCreated: useCallback((newChatId: string) => {
      console.log('[ChatPage] onChatCreated called:', { newChatId, currentSelectedChatId: selectedChatId });
      setSelectedChatId(newChatId);
      reloadChats();
    }, [reloadChats, selectedChatId]),
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

  useEffect(() => {
    if (!loading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [loading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

  // Инициализация при первой загрузке
  useEffect(() => {
    const loadInitialData = async () => {
      if (!currentOrg?.id || !isAuthenticated || isNewUser || chatsLoading) return;

      try {
        const limitsResp = await agent.v1.agentServiceGetLlmLimits();
        setLimits(limitsResp.data ?? null);

        // Устанавливаем новый чат только при первой загрузке
        setSelectedChatId(null);
        setInitialLoading(false);
      } catch (e) {
        console.error("Failed to load initial data", e);
        setInitialLoading(false);
      }
    };

    // Выполняем только один раз при монтировании компонента
    if (initialLoading && !chatsLoading) {
      void loadInitialData();
    }
  }, [agent.v1, currentOrg?.id, isAuthenticated, isNewUser, chatsLoading, initialLoading]);

  const handleSelectChat = useCallback(
    async (chatId: string) => {
      setSelectedChatId(chatId);
      clearError();
      setLoadingMessages(true);

      try {
        // Загружаем сообщения
        const messagesResp = await agent.v1.agentServiceGetMessages(chatId, { 
          orgId: currentOrg?.id ?? "",
          limit: 50, 
          offset: 0 
        });
        setMessages(messagesResp.data.messages ?? []);
        
        // Загружаем информацию о чате
        const chatResp = await agent.v1.agentServiceGetChat(chatId, { orgId: currentOrg?.id ?? "" });
        if (chatResp.data.chat?.title) {
          setChatName(chatResp.data.chat.title);
        }
      } catch (e) {
        console.error("Failed to load chat data", e);
      } finally {
        setLoadingMessages(false);
      }
    },
    [agent.v1, currentOrg?.id, setMessages, setChatName, clearError]
  );

  const handleCreateChat = useCallback(() => {
    setSelectedChatId(null);
    setMessages([]);
    setChatName(null);
    clearError();
  }, [setMessages, setChatName, clearError]);

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
      <Card className="flex flex-1 flex-col rounded-3xl shadow-none">
        <ChatHeader
          chatName={chatName}
          limits={limits}
          usageTokens={usageTokens}
          onShowChatList={() => setShowChatListModal(true)}
        />

        <ChatWindow 
          messages={messages} 
          streamingMessage={streamingMessage} 
          streamingToolCalls={streamingToolCalls}
          isStreaming={isStreaming}
          loadingMessages={loadingMessages}
        />

        <ChatInput
          value={input}
          onChange={setInput}
          onSend={handleSend}
          disabled={false}
          isStreaming={isStreaming}
        />
      </Card>

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
