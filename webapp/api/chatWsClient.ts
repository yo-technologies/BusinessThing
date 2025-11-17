"use client";

import { AgentStreamMessageResponse } from "@/api/api.agent.generated";
import { getAuthToken } from "@/api/client";

export type ChatWsEvent = AgentStreamMessageResponse;

export type ChatWsCallbacks = {
  onMessage?: (event: ChatWsEvent) => void;
  onError?: (error: Event) => void;
  onClose?: () => void;
};

export const createChatWebSocket = (
  chatId: string | null,
  orgId?: string,
  callbacks?: ChatWsCallbacks,
) => {
  const token = getAuthToken();
  const params = new URLSearchParams();

  if (chatId !== null) {
    params.set("chatId", chatId);
  }
  if (orgId) params.set("orgId", orgId);
  if (token) params.set("token", token);

  const wsUrl = `wss://agent.businessthing.ru/api/ws/chat?${params.toString()}`;

  const socket = new WebSocket(wsUrl);

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as ChatWsEvent;

      callbacks?.onMessage?.(data);
    } catch (e) {
      console.error(
        "[chatWsClient] Failed to parse WS message:",
        e,
        "Raw data:",
        event.data,
      );
    }
  };

  socket.onerror = (event) => {
    console.error("[chatWsClient] WebSocket error:", {
      event,
      readyState: socket.readyState,
      url: socket.url.replace(/token=[^&]+/, "token=***"),
    });
    callbacks?.onError?.(event);
  };

  socket.onclose = (event) => {
    const closeReasons: Record<number, string> = {
      1000: "Normal Closure",
      1001: "Going Away",
      1002: "Protocol Error",
      1003: "Unsupported Data",
      1006: "Abnormal Closure (no close frame)",
      1007: "Invalid frame payload data",
      1008: "Policy Violation",
      1009: "Message too big",
      1011: "Internal Server Error",
      1015: "TLS Handshake Failed",
    };

    callbacks?.onClose?.();
  };

  return socket;
};
