"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AgentMessage, AgentMessageRole } from "@/api/api.agent.generated";
import { createChatWebSocket, ChatWsEvent } from "@/api/chatWsClient";

interface UseChatWebSocketParams {
  chatId: string | null;
  organizationId?: string;
  onMessageReceived?: (message: AgentMessage) => void;
  onChatCreated?: (chatId: string) => void;
}

interface ChatWebSocketState {
  messages: AgentMessage[];
  streamingMessage: string;
  isStreaming: boolean;
  error: string | null;
  usageTokens: number | null;
}

export function useChatWebSocket({ chatId, organizationId, onMessageReceived, onChatCreated }: UseChatWebSocketParams) {
  const wsRef = useRef<WebSocket | null>(null);
  const [currentChatId, setCurrentChatId] = useState<string | null>(chatId);
  const [state, setState] = useState<ChatWebSocketState>({
    messages: [],
    streamingMessage: "",
    isStreaming: false,
    error: null,
    usageTokens: null,
  });

  useEffect(() => {
    setCurrentChatId(chatId);
  }, [chatId]);

  useEffect(() => {
    wsRef.current?.close();

    console.log('[useChatWebSocket] Establishing WebSocket connection:', {
      currentChatId,
      organizationId,
    });

    wsRef.current = createChatWebSocket(currentChatId, organizationId, {
      onMessage: (event: ChatWsEvent) => {
        console.log('[useChatWebSocket] Received event:', event);

        if (event.usage?.usage?.totalTokens) {
          console.log('[useChatWebSocket] Usage update:', event.usage.usage.totalTokens);
          setState((prev) => ({
            ...prev,
            usageTokens: event.usage?.usage?.totalTokens ?? null,
          }));
        }

        if (event.chunk?.content) {
          console.log('[useChatWebSocket] Received chunk:', event.chunk.content.substring(0, 20) + '...');
          setState((prev) => ({
            ...prev,
            isStreaming: true,
            streamingMessage: prev.streamingMessage + (event.chunk?.content ?? ""),
          }));
        }

        if (event.message) {
          console.log('[useChatWebSocket] Received complete message:', {
            messageId: event.message.id,
            role: event.message.role,
            chatId: event.message.chatId,
          });
          setState((prev) => ({
            ...prev,
            messages: [...prev.messages, event.message!],
            streamingMessage: "",
            isStreaming: false,
          }));
          onMessageReceived?.(event.message);
          
          // Если получили сообщение с chatId, а текущий chatId null - значит чат создан
          if (!currentChatId && event.message.chatId) {
            console.log('[useChatWebSocket] New chat created:', event.message.chatId);
            setCurrentChatId(event.message.chatId);
            onChatCreated?.(event.message.chatId);
          }
        }

        if (event.error) {
          console.error('[useChatWebSocket] Received error:', event.error);
          setState((prev) => ({
            ...prev,
            error: event.error?.message ?? "Ошибка генерации ответа",
            isStreaming: false,
          }));
        }
      },
      onError: (error) => {
        console.error('[useChatWebSocket] WebSocket error:', error);
        setState((prev) => ({
          ...prev,
          error: "Проблема с подключением к чату",
        }));
      },
      onClose: () => {
        console.log('[useChatWebSocket] WebSocket closed');
        setState((prev) => ({
          ...prev,
          isStreaming: false,
        }));
      },
    });

    return () => {
      console.log('[useChatWebSocket] Cleaning up WebSocket connection');
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [currentChatId, organizationId, onMessageReceived, onChatCreated]);

  const sendMessage = useCallback(
    (content: string) => {
      if (!wsRef.current) {
        console.error('[useChatWebSocket] Cannot send message: WebSocket not connected');
        return;
      }

      console.log('[useChatWebSocket] Preparing to send message:', {
        content: content.substring(0, 50) + '...',
        currentChatId,
        organizationId,
        wsReadyState: wsRef.current.readyState,
      });

      const optimisticMessage: AgentMessage = {
        id: `${Date.now()}`,
        chatId: currentChatId || "",
        role: AgentMessageRole.MESSAGE_ROLE_USER,
        content,
        createdAt: new Date().toISOString(),
      };

      setState((prev) => ({
        ...prev,
        messages: [...prev.messages, optimisticMessage],
        error: null,
      }));

      // Формируем payload в соответствии с protobuf контрактом StreamMessageRequest
      const payload: any = {
        newMessage: {
          content,
          orgId: organizationId,
        },
      };
      
      // Добавляем chatId только если он есть
      if (currentChatId) {
        payload.newMessage.chatId = currentChatId;
      }

      console.log('[useChatWebSocket] Sending payload:', JSON.stringify(payload));

      try {
        wsRef.current.send(JSON.stringify(payload));
        console.log('[useChatWebSocket] Message sent successfully');
        setState((prev) => ({
          ...prev,
          isStreaming: true,
          streamingMessage: "",
        }));
      } catch (e) {
        console.error('[useChatWebSocket] Failed to send message:', e);
        setState((prev) => ({
          ...prev,
          error: "Не удалось отправить сообщение",
          isStreaming: false,
        }));
      }
    },
    [currentChatId, organizationId]
  );

  const setMessages = useCallback((messages: AgentMessage[]) => {
    setState((prev) => ({ ...prev, messages }));
  }, []);

  const clearError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  return {
    ...state,
    sendMessage,
    setMessages,
    clearError,
  };
}
