"use client";

import { AgentStreamMessageResponse } from "@/api/api.agent.generated";
import { getAuthToken } from "@/api/client";

export type ChatWsEvent = AgentStreamMessageResponse;

export type ChatWsCallbacks = {
  onMessage?: (event: ChatWsEvent) => void;
  onError?: (error: Event) => void;
  onClose?: () => void;
};

export const createChatWebSocket = (chatId: string, orgId?: string, callbacks?: ChatWsCallbacks) => {
  const token = getAuthToken();
  const params = new URLSearchParams();

  params.set("chatId", chatId);
  if (orgId) params.set("orgId", orgId);
  if (token) params.set("token", token);

  const wsUrl = `wss://agent.businessthing.ru/api/ws/chat?${params.toString()}`;

  const socket = new WebSocket(wsUrl);

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as ChatWsEvent;
      callbacks?.onMessage?.(data);
    } catch (e) {
      console.error("Failed to parse WS message", e);
    }
  };

  socket.onerror = (event) => {
    callbacks?.onError?.(event);
  };

  socket.onclose = () => {
    callbacks?.onClose?.();
  };

  return socket;
};
