"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AgentMessage, AgentMessageRole, AgentToolCallEvent } from "@/api/api.agent.generated";
import { createChatWebSocket, ChatWsEvent } from "@/api/chatWsClient";

interface UseChatWebSocketParams {
  chatId: string | null;
  organizationId?: string;
  onMessageReceived?: (message: AgentMessage) => void;
  onChatCreated?: (chatId: string) => void;
  onFinalReceived?: () => void;
}

interface ChatWebSocketState {
  messages: AgentMessage[];
  streamingMessage: string;
  streamingToolCalls: Map<string, AgentToolCallEvent>;
  isStreaming: boolean;
  error: string | null;
  usageTokens: number | null;
  chatName: string | null;
}

export function useChatWebSocket({ chatId, organizationId, onMessageReceived, onChatCreated, onFinalReceived }: UseChatWebSocketParams) {
  const wsRef = useRef<WebSocket | null>(null);
  const [currentChatId, setCurrentChatId] = useState<string | null>(chatId);
  const [state, setState] = useState<ChatWebSocketState>({
    messages: [],
    streamingMessage: "",
    streamingToolCalls: new Map(),
    isStreaming: false,
    error: null,
    usageTokens: null,
    chatName: null,
  });

  useEffect(() => {
    setCurrentChatId(chatId);
    // Сбрасываем состояние при смене chatId
    setState({
      messages: [],
      streamingMessage: "",
      streamingToolCalls: new Map(),
      isStreaming: false,
      error: null,
      usageTokens: null,
      chatName: null,
    });
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

        if (event.toolCall) {
          console.log('[useChatWebSocket] Received tool call event:', event.toolCall);
          setState((prev) => {
            const newToolCalls = new Map(prev.streamingToolCalls);
            const toolCallId = event.toolCall?.toolCallId ?? "";
            
            if (toolCallId) {
              newToolCalls.set(toolCallId, event.toolCall!);
            }
            
            return {
              ...prev,
              streamingToolCalls: newToolCalls,
            };
          });
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
            streamingToolCalls: new Map(),
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

        if (event.chat) {
          console.log('[useChatWebSocket] Received chat event:', event.chat);
          if (event.chat.chatId && event.chat.chatId !== currentChatId) {
            console.log('[useChatWebSocket] Updating chat ID:', event.chat.chatId);
            setCurrentChatId(event.chat.chatId);
            onChatCreated?.(event.chat.chatId);
          }
          if (event.chat.chatName) {
            console.log('[useChatWebSocket] Updating chat name:', event.chat.chatName);
            setState((prev) => ({
              ...prev,
              chatName: event.chat?.chatName ?? null,
            }));
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

        if (event.final) {
          console.log('[useChatWebSocket] Received final event:', event.final);
          // Final событие содержит полное финальное состояние чата - используем его как источник правды
          setState((prev) => ({
            ...prev,
            messages: event.final?.messages ?? prev.messages,
            streamingMessage: "",
            streamingToolCalls: new Map(),
            isStreaming: false,
            usageTokens: 0, // Сбрасываем usage токены для следующего сообщения
          }));
          
          // Обновляем chatName если он есть в final.chat
          if (event.final.chat?.title) {
            setState((prev) => ({
              ...prev,
              chatName: event.final?.chat?.title ?? prev.chatName,
            }));
          }
          
          // Уведомляем о завершении стрима для перезагрузки лимитов
          onFinalReceived?.();
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
  }, [currentChatId, organizationId, onMessageReceived, onChatCreated, onFinalReceived]);

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
          streamingToolCalls: new Map(),
        }));      // Формируем payload в соответствии с protobuf контрактом StreamMessageRequest
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
    setState((prev) => ({ ...prev, messages, streamingToolCalls: new Map() }));
  }, []);

  const setChatName = useCallback((chatName: string | null) => {
    setState((prev) => ({ ...prev, chatName }));
  }, []);

  const clearError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  return {
    ...state,
    currentChatId,
    sendMessage,
    setMessages,
    setChatName,
    clearError,
  };
}
