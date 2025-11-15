"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Input } from "@heroui/input";
import { Spinner } from "@heroui/spinner";
import { Divider } from "@heroui/divider";
import { useRouter } from "next/navigation";

import { AgentChat, AgentGetLLMLimitsResponse, AgentMessage, AgentMessageRole } from "@/api/api.agent.generated";
import { useApiClients } from "@/api/client";
import { createChatWebSocket } from "@/api/chatWsClient";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";

type StreamState = {
  status?: string;
  usageTokens?: number;
};

export default function ChatPage() {
  const router = useRouter();
  const { isAuthenticated, loading, user, isNewUser } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization();
  const { agent } = useApiClients();

  const [chat, setChat] = useState<AgentChat | null>(null);
  const [messages, setMessages] = useState<AgentMessage[]>([]);
  const [input, setInput] = useState("");
  const [streamingAssistantMessage, setStreamingAssistantMessage] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamState, setStreamState] = useState<StreamState>({});
  const [limits, setLimits] = useState<AgentGetLLMLimitsResponse | null>(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const scrollRef = useRef<HTMLDivElement | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const chatId = useMemo(() => chat?.id ?? null, [chat?.id]);

  useEffect(() => {
    if (!loading && isAuthenticated && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isAuthenticated, isNewUser, loading, router]);

  useEffect(() => {
    if (!loading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [loading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

  const scrollToBottom = useCallback(() => {
    if (!scrollRef.current) return;
    try {
      scrollRef.current.scrollTo({ top: scrollRef.current.scrollHeight });
    } catch {
      // no-op
    }
  }, []);

  const loadChatAndLimits = useCallback(async () => {
    setInitialLoading(true);
    setError(null);
    try {
      const [chatsResp, limitsResp] = await Promise.all([
        agent.v1.agentServiceListChats({ page: 1, pageSize: 1 }),
        agent.v1.agentServiceGetLlmLimits(),
      ]);

      const firstChat = chatsResp.data.chats?.[0] ?? null;
      setChat(firstChat);
      setLimits(limitsResp.data ?? null);

      if (firstChat?.id) {
        const messagesResp = await agent.v1.agentServiceGetMessages(firstChat.id, { limit: 50, offset: 0 });
        setMessages(messagesResp.data.messages ?? []);
      } else {
        setMessages([]);
      }
    } catch (e) {
      console.error("Failed to load chat or limits", e);
      setError("Не удалось загрузить чат");
    } finally {
      setInitialLoading(false);
      setTimeout(scrollToBottom, 0);
    }
  }, [agent.v1, scrollToBottom]);

  useEffect(() => {
    if (!isAuthenticated || loading || isNewUser || !currentOrg?.id) return;
    void loadChatAndLimits();
  }, [isAuthenticated, isNewUser, loading, currentOrg?.id, loadChatAndLimits]);

  useEffect(() => {
    if (!chatId) return;

    wsRef.current?.close();

    wsRef.current = createChatWebSocket(chatId, chat?.organizationId, {
      onMessage: (event) => {
        if (event.usage?.usage?.totalTokens) {
          setStreamState({
            usageTokens: event.usage.usage.totalTokens,
            status: event.error ? event.error.message : undefined,
          });
        }

        if (event.chunk?.content) {
          setIsStreaming(true);
          setStreamingAssistantMessage((prev) => prev + (event.chunk?.content ?? ""));
          setTimeout(scrollToBottom, 0);
        }

        if (event.message) {
          setMessages((prev) => [...prev, event.message!]);
          setStreamingAssistantMessage("");
          setIsStreaming(false);
          setTimeout(scrollToBottom, 0);
        }

        if (event.error) {
          setError(event.error.message ?? "Ошибка генерации ответа");
          setIsStreaming(false);
        }
      },
      onError: () => {
        setError("Проблема с подключением к чату");
      },
      onClose: () => {
        setIsStreaming(false);
      },
    });

    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [chat?.organizationId, chatId, scrollToBottom]);

  const handleSend = () => {
    if (!input.trim() || !chatId || !wsRef.current) return;

    const content = input.trim();
    setInput("");
    setError(null);

    const optimisticMessage: AgentMessage = {
      id: `${Date.now()}`,
      chatId,
      role: AgentMessageRole.MESSAGE_ROLE_USER,
      content,
      createdAt: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimisticMessage]);
    setTimeout(scrollToBottom, 0);

    const payload = {
      type: "new_message",
      chatId,
      content,
      orgId: chat?.organizationId,
    };

    try {
      wsRef.current.send(JSON.stringify(payload));
      setIsStreaming(true);
      setStreamingAssistantMessage("");
    } catch (e) {
      console.error("Failed to send WS message", e);
      setError("Не удалось отправить сообщение");
      setIsStreaming(false);
    }
  };

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
        <p className="text-small text-default-500 text-center">
          Попробуй закрыть мини-приложение и открыть его заново из Telegram.
        </p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-3">
      <Card className="border-none bg-default-50/60 shadow-sm">
        <CardHeader className="flex items-center justify-between gap-2 pb-2">
          <div className="flex flex-col items-start gap-1">
            <span className="text-tiny font-medium uppercase text-primary">Чат с агентом</span>
            <h1 className="text-xl font-semibold">{user?.firstName ?? "Клиент"}</h1>
          </div>
          {limits && (
            <div className="flex flex-col items-end text-right text-[11px] text-default-500">
              <span>
                Сегодня: {limits.used ?? 0} / {limits.dailyLimit ?? 0} токенов
              </span>
              <span>Осталось: {limits.remaining ?? 0}</span>
            </div>
          )}
        </CardHeader>
        <CardBody className="space-y-3 pb-4">
          <p className="text-small text-default-500">
            Задавай вопросы по документам, а агент поможет с разбором и подготовкой договоров.
          </p>
          {error && (
            <p className="text-xs text-danger-500">{error}</p>
          )}
          {streamState.usageTokens && (
            <p className="text-[11px] text-default-400">
              Потрачено токенов: {streamState.usageTokens}
            </p>
          )}
        </CardBody>
      </Card>

      <Card className="flex min-h-0 flex-1 flex-col border-none bg-content1/70 shadow-sm">
        <CardBody className="flex flex-1 flex-col gap-3 overflow-y-auto pb-4">
          {messages.length === 0 && !streamingAssistantMessage ? (
            <p className="mt-4 text-center text-small text-default-400">
              Пока сообщений нет. Напиши что-нибудь, чтобы начать диалог.
            </p>
          ) : (
            <>
              {messages.map((message) => {
                const isUser = message.role === AgentMessageRole.MESSAGE_ROLE_USER;
                return (
                  <div
                    key={message.id}
                    className={
                      isUser
                        ? "ml-auto max-w-[80%] rounded-xl bg-primary text-primary-foreground px-3 py-2 text-small"
                        : "mr-auto max-w-[80%] rounded-xl bg-default-100 px-3 py-2 text-small"
                    }
                  >
                    {message.content}
                  </div>
                );
              })}
              {streamingAssistantMessage && (
                <div className="mr-auto max-w-[80%] rounded-xl bg-default-100 px-3 py-2 text-small">
                  {streamingAssistantMessage}
                  {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
                </div>
              )}
            </>
          )}
        </CardBody>
      </Card>

      <Divider className="opacity-40" />

      <div className="flex items-end gap-2">
        <Input
          size="lg"
          radius="lg"
          variant="bordered"
          placeholder="Напиши сообщение..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              handleSend();
            }
          }}
        />
        <Button color="primary" radius="lg" onPress={handleSend} isDisabled={!input.trim() || !chatId}>
          {isStreaming ? "Генерируем..." : "Отправить"}
        </Button>
      </div>
    </div>
  );
}

