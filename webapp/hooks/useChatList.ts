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

  useEffect(() => {
    void loadChats();
  }, [loadChats]);

  return {
    chats,
    loading,
    error,
    reload: loadChats,
  };
}
