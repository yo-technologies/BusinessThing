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

export function useChatWebSocket({ 
  chatId, 
  organizationId, 
  onMessageReceived, 
  onChatCreated, 
  onFinalReceived 
}: UseChatWebSocketParams) {
  const wsRef = useRef<WebSocket | null>(null);
  const activeChatIdRef = useRef<string | null>(chatId);
  
  const callbacksRef = useRef({ onMessageReceived, onChatCreated, onFinalReceived });
  
  useEffect(() => {
    callbacksRef.current = { onMessageReceived, onChatCreated, onFinalReceived };
  });
  
  const [state, setState] = useState<ChatWebSocketState>({
    messages: [],
    streamingMessage: "",
    streamingToolCalls: new Map(),
    isStreaming: false,
    error: null,
    usageTokens: null,
    chatName: null,
  });

  // Сбрасываем состояние при смене чата
  useEffect(() => {
    console.log('[useChatWebSocket] ChatId changed:', { 
      from: activeChatIdRef.current, 
      to: chatId 
    });
    
    // Если chatId не изменился - ничего не делаем
    if (activeChatIdRef.current === chatId) {
      console.log('[useChatWebSocket] ChatId unchanged, skipping');
      return;
    }
    
    console.log('[useChatWebSocket] Resetting state for new chatId');
    
    // Закрываем WebSocket при смене чата
    if (wsRef.current) {
      console.log('[useChatWebSocket] Closing WebSocket due to chatId change');
      wsRef.current.close();
      wsRef.current = null;
    }
    
    activeChatIdRef.current = chatId;
    
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

  // Функция для создания WebSocket соединения
  const createWebSocketConnection = useCallback(() => {
    if (!organizationId) {
      console.error('[useChatWebSocket] Cannot create WebSocket: no organizationId');
      return null;
    }

    if (wsRef.current) {
      console.log('[useChatWebSocket] WebSocket already exists');
      return wsRef.current;
    }

    console.log('[useChatWebSocket] Creating WebSocket connection:', {
      chatId: activeChatIdRef.current,
      organizationId,
    });

    const ws = createChatWebSocket(activeChatIdRef.current, organizationId, {
      onMessage: (event: ChatWsEvent) => {
        // Обработка usage
        if (event.usage?.usage?.totalTokens) {
          setState((prev) => ({
            ...prev,
            usageTokens: event.usage?.usage?.totalTokens ?? null,
          }));
        }

        // Обработка стриминга контента
        if (event.chunk?.content) {
          setState((prev) => ({
            ...prev,
            isStreaming: true,
            streamingMessage: prev.streamingMessage + (event.chunk?.content ?? ""),
          }));
        }

        // Обработка tool calls
        if (event.toolCall) {
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

        // Обработка полного сообщения
        if (event.message) {
          setState((prev) => ({
            ...prev,
            messages: [...prev.messages, event.message!],
            streamingMessage: "",
            streamingToolCalls: new Map(),
            isStreaming: false,
          }));
          
          callbacksRef.current.onMessageReceived?.(event.message);
          
          // Если это новый чат - обновляем активный chatId
          if (activeChatIdRef.current === null && event.message.chatId) {
            console.log('[useChatWebSocket] New chat created:', event.message.chatId);
            activeChatIdRef.current = event.message.chatId;
            callbacksRef.current.onChatCreated?.(event.message.chatId);
          }
        }

        // Обработка события чата
        if (event.chat) {
          // Обновляем активный chatId если это новый чат
          if (activeChatIdRef.current === null && event.chat.chatId) {
            console.log('[useChatWebSocket] New chat created from chat event:', event.chat.chatId);
            activeChatIdRef.current = event.chat.chatId;
            callbacksRef.current.onChatCreated?.(event.chat.chatId);
          }
          
          // Обновляем название чата
          if (event.chat.chatName) {
            setState((prev) => ({
              ...prev,
              chatName: event.chat?.chatName ?? null,
            }));
          }
        }

        // Обработка ошибок
        if (event.error) {
          console.error('[useChatWebSocket] Received error:', event.error);
          setState((prev) => ({
            ...prev,
            error: event.error?.message ?? "Ошибка генерации ответа",
            isStreaming: false,
          }));
        }

        // Обработка финального события
        if (event.final) {
          setState((prev) => ({
            ...prev,
            messages: event.final?.messages ?? prev.messages,
            streamingMessage: "",
            streamingToolCalls: new Map(),
            isStreaming: false,
            usageTokens: 0,
            chatName: event.final?.chat?.title ?? prev.chatName,
          }));
          
          callbacksRef.current.onFinalReceived?.();
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
        wsRef.current = null;
      },
    });

    wsRef.current = ws;
    return ws;
  }, [organizationId]);

  const sendMessage = useCallback(
    (content: string) => {
      // Создаем WebSocket при отправке сообщения, если его еще нет
      let ws = wsRef.current;
      
      if (!ws) {
        console.log('[useChatWebSocket] Creating WebSocket for message send');
        ws = createWebSocketConnection();
        
        if (!ws) {
          setState((prev) => ({
            ...prev,
            error: "Не удалось установить соединение",
          }));
          return;
        }
      }

      // Ждем открытия соединения если оно еще не открыто
      if (ws.readyState === WebSocket.CONNECTING) {
        console.log('[useChatWebSocket] Waiting for WebSocket to open...');
        ws.addEventListener('open', () => {
          sendMessageInternal(ws!, content);
        }, { once: true });
      } else if (ws.readyState === WebSocket.OPEN) {
        sendMessageInternal(ws, content);
      } else {
        console.error('[useChatWebSocket] WebSocket is not in a valid state:', ws.readyState);
        setState((prev) => ({
          ...prev,
          error: "Соединение не установлено",
        }));
      }
    },
    [organizationId]
  );

  const sendMessageInternal = useCallback((ws: WebSocket, content: string) => {
    const optimisticMessage: AgentMessage = {
      id: `temp-${Date.now()}`,
      chatId: activeChatIdRef.current || "",
      role: AgentMessageRole.MESSAGE_ROLE_USER,
      content,
      createdAt: new Date().toISOString(),
    };

    setState((prev) => ({
      ...prev,
      messages: [...prev.messages, optimisticMessage],
      error: null,
      streamingToolCalls: new Map(),
    }));

    const payload: any = {
      newMessage: {
        content,
        orgId: organizationId,
      },
    };
    
    // Добавляем chatId только если он есть
    if (activeChatIdRef.current) {
      payload.newMessage.chatId = activeChatIdRef.current;
    }

    try {
      ws.send(JSON.stringify(payload));
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
  }, [organizationId]);

  const setMessages = useCallback((messages: AgentMessage[]) => {
    setState((prev) => ({ ...prev, messages, streamingToolCalls: new Map() }));
  }, []);

  const setChatName = useCallback((chatName: string | null) => {
    setState((prev) => ({ ...prev, chatName }));
  }, []);

  const clearError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  // Cleanup при размонтировании
  useEffect(() => {
    return () => {
      console.log('[useChatWebSocket] Component unmounting, closing WebSocket');
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, []);

  return {
    ...state,
    currentChatId: activeChatIdRef.current,
    sendMessage,
    setMessages,
    setChatName,
    clearError,
  };
}
