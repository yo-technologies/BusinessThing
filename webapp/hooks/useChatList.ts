"use client";

import { useCallback, useEffect, useState } from "react";
import { AgentChat } from "@/api/api.agent.generated";
import { useApiClients } from "@/api/client";

interface UseChatListParams {
  organizationId?: string;
  enabled?: boolean;
}

export function useChatList({ organizationId, enabled = true }: UseChatListParams) {
  const { agent } = useApiClients();
  const [chats, setChats] = useState<AgentChat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadChats = useCallback(async () => {
    if (!organizationId || !enabled) return;

    setLoading(true);
    setError(null);

    try {
      const response = await agent.v1.agentServiceListChats({
        orgId: organizationId,
        page: 1,
        pageSize: 50,
      });

      setChats(response.data.chats ?? []);
    } catch (e) {
      console.error("Failed to load chats", e);
      setError("Не удалось загрузить список чатов");
    } finally {
      setLoading(false);
    }
  }, [agent.v1, organizationId, enabled]);

  const deleteChat = useCallback(async (chatId: string) => {
    if (!organizationId) return;

    try {
      await agent.v1.agentServiceDeleteChat(chatId, { orgId: organizationId });
      
      setChats((prevChats) => prevChats.filter((chat) => chat.id !== chatId));
    } catch (e) {
      console.error("Failed to delete chat", e);
      setError("Не удалось удалить чат");
      throw e;
    }
  }, [agent.v1, organizationId]);

  useEffect(() => {
    void loadChats();
  }, [loadChats]);

  return {
    chats,
    loading,
    error,
    reload: loadChats,
    deleteChat,
  };
}
