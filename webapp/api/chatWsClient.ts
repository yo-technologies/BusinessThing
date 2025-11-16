"use client";

import { AgentStreamMessageResponse } from "@/api/api.agent.generated";
import { getAuthToken } from "@/api/client";

export type ChatWsEvent = AgentStreamMessageResponse;

export type ChatWsCallbacks = {
  onMessage?: (event: ChatWsEvent) => void;
  onError?: (error: Event) => void;
  onClose?: () => void;
};

export const createChatWebSocket = (chatId: string | null, orgId?: string, callbacks?: ChatWsCallbacks) => {
  const token = getAuthToken();
  const params = new URLSearchParams();

  console.log('[chatWsClient] Auth token details:', {
    available: !!token,
    length: token?.length,
    prefix: token?.substring(0, 10),
    hasBearer: token?.startsWith('Bearer '),
    hasTma: token?.startsWith('tma ')
  });

  if (chatId !== null) {
    params.set("chatId", chatId);
  }
  if (orgId) params.set("orgId", orgId);
  if (token) params.set("token", token);

  const wsUrl = `wss://agent.businessthing.ru/api/ws/chat?${params.toString()}`;

  console.log('[chatWsClient] Creating WebSocket connection:', { 
    wsUrl: wsUrl.replace(/token=[^&]+/, 'token=***'), 
    chatId, 
    orgId,
    hasToken: !!token,
    queryParams: Object.fromEntries(params.entries()).token ? '...with token' : 'no token'
  });

  const socket = new WebSocket(wsUrl);

  socket.onopen = () => {
    console.log('[chatWsClient] WebSocket connection opened');
  };

  socket.onmessage = (event) => {
    console.log('[chatWsClient] Raw message received:', event.data);
    try {
      const data = JSON.parse(event.data) as ChatWsEvent;
      console.log('[chatWsClient] Parsed message:', data);
      callbacks?.onMessage?.(data);
    } catch (e) {
      console.error('[chatWsClient] Failed to parse WS message:', e, 'Raw data:', event.data);
    }
  };

  socket.onerror = (event) => {
    console.error('[chatWsClient] WebSocket error:', {
      event,
      readyState: socket.readyState,
      url: socket.url.replace(/token=[^&]+/, 'token=***')
    });
    callbacks?.onError?.(event);
  };

  socket.onclose = (event) => {
    const closeReasons: Record<number, string> = {
      1000: 'Normal Closure',
      1001: 'Going Away',
      1002: 'Protocol Error',
      1003: 'Unsupported Data',
      1006: 'Abnormal Closure (no close frame)',
      1007: 'Invalid frame payload data',
      1008: 'Policy Violation',
      1009: 'Message too big',
      1011: 'Internal Server Error',
      1015: 'TLS Handshake Failed'
    };
    console.log('[chatWsClient] WebSocket closed:', { 
      code: event.code, 
      reason: event.reason || closeReasons[event.code] || 'Unknown',
      wasClean: event.wasClean,
      url: socket.url.replace(/token=[^&]+/, 'token=***')
    });
    callbacks?.onClose?.();
  };

  return socket;
};
