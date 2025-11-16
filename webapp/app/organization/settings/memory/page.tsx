"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { TrashIcon, LightBulbIcon } from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { AgentMemoryFact } from "@/api/api.agent.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";

type FactWithLoading = AgentMemoryFact & { deleting?: boolean };

export default function MemoryPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations, authLoading });
  const { agent } = useApiClients();

  const [facts, setFacts] = useState<FactWithLoading[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useBackButton(true);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, authLoading, router]);

  useEffect(() => {
    if (!authLoading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [authLoading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

  const loadFacts = useCallback(async () => {
    if (!currentOrg?.id) return;

    setInitialLoading(true);
    setError(null);
    try {
      const response = await agent.v1.memoryServiceListMemoryFacts({ orgId: currentOrg.id });
      setFacts(response.data.facts ?? []);
    } catch (e) {
      console.error("Failed to load memory facts", e);
      setError("Не удалось загрузить память агента");
    } finally {
      setInitialLoading(false);
    }
  }, [agent.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadFacts();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadFacts]);

  const handleDelete = async (fact: AgentMemoryFact) => {
    if (!fact.id || !currentOrg?.id) return;

    setFacts((prev) =>
      prev.map((f) => (f.id === fact.id ? { ...f, deleting: true } : f)),
    );

    try {
      await agent.v1.memoryServiceDeleteMemoryFact(fact.id, { orgId: currentOrg.id });
      setFacts((prev) => prev.filter((f) => f.id !== fact.id));
    } catch (e) {
      console.error("Failed to delete fact", e);
      setFacts((prev) =>
        prev.map((f) => (f.id === fact.id ? { ...f, deleting: false } : f)),
      );
    }
  };

  if (authLoading || orgLoading || initialLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <Card className="max-w-xl mx-auto mt-8 rounded-4xl shadow-sm border border-danger-200/60">
        <CardBody>
          <p className="text-danger text-center">{error}</p>
        </CardBody>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4 flex-1">
      <Card className="rounded-4xl shadow-none">
        <CardHeader className="flex flex-col gap-1 px-5 py-4">
          <div className="flex items-center gap-2 w-full">
            <LightBulbIcon className="h-6 w-6 flex-shrink-0 text-warning" />
            <p className="text-xl font-semibold">Память агента</p>
          </div>
          <p className="text-xs text-default-300">
            Факты, которые агент будет использовать при общении с пользователями вашей организации.
          </p>
        </CardHeader>
      </Card>

      <Card className="flex-1 rounded-4xl shadow-none h-full">
        <CardBody className="gap-2">
          {facts.length === 0 ? (
            <div className="flex flex-col h-full items-center justify-center py-12 gap-2">
              <p className="text-default-400 text-center">Память агента пуста</p>
              <p className="text-sm text-default-300 text-center max-w-md">
                Общайтесь с вашим агентом в чате, и важные факты будут автоматически сохраняться здесь.
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {facts.map((fact) => (
                <Card
                  key={fact.id}
                  className="flex flex-col gap-3 p-4 rounded-3xl bg-default-50/60 hover:bg-default-100/70 transition-colors"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm whitespace-pre-wrap">{fact.content}</p>
                  </div>
                  <div className="flex gap-2 justify-end">
                    <Button
                      isIconOnly
                      size="sm"
                      variant="flat"
                      color="danger"
                      isLoading={fact.deleting}
                      onPress={() => handleDelete(fact)}
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                </Card>
              ))}
            </div>
          )}
        </CardBody>
      </Card>

    </div>
  );
}
