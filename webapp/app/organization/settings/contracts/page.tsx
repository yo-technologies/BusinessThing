"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Modal, ModalBody, ModalContent, ModalFooter, ModalHeader, useDisclosure } from "@heroui/modal";
import {
  DocumentTextIcon,
  TrashIcon,
  ArrowDownTrayIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { CoreGeneratedContract } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";

type ContractWithLoading = CoreGeneratedContract & { deleting?: boolean };

export default function ContractsPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations, authLoading });
  const { core } = useApiClients();

  const [contracts, setContracts] = useState<ContractWithLoading[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedContract, setSelectedContract] = useState<CoreGeneratedContract | null>(null);

  const { isOpen, onOpen, onClose } = useDisclosure();

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

  const loadContracts = useCallback(async () => {
    if (!currentOrg?.id) return;

    setInitialLoading(true);
    setError(null);
    try {
      const response = await core.v1.generatedContractServiceListContracts(currentOrg.id, {
        pageSize: 100,
      });
      setContracts(response.data.contracts ?? []);
    } catch (e) {
      console.error("Failed to load contracts", e);
      setError("Не удалось загрузить документы");
    } finally {
      setInitialLoading(false);
    }
  }, [core.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadContracts();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadContracts]);

  const handleDelete = async (contract: CoreGeneratedContract) => {
    if (!contract.id) return;

    setContracts((prev) =>
      prev.map((c) => (c.id === contract.id ? { ...c, deleting: true } : c)),
    );

    try {
      await core.v1.generatedContractServiceDeleteContract(contract.id);
      setContracts((prev) => prev.filter((c) => c.id !== contract.id));
      onClose();
    } catch (e) {
      console.error("Failed to delete contract", e);
      setContracts((prev) =>
        prev.map((c) => (c.id === contract.id ? { ...c, deleting: false } : c)),
      );
    }
  };

  const handleDownload = async (contract: CoreGeneratedContract) => {
    if (!contract.s3Key) return;

    // TODO: Implement download after storage API is generated
    console.warn("Download not implemented yet, s3Key:", contract.s3Key);
  };

  const openDeleteModal = (contract: CoreGeneratedContract) => {
    setSelectedContract(contract);
    onOpen();
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
          <div className="flex items-start gap-2 w-full">
            <DocumentTextIcon className="h-5 w-5 flex-shrink-0 text-success mt-1" />
            <p className="text-xl font-semibold">Сгенерированные документы</p>
          </div>
          <p className="text-xs text-default-300">
            История контрактов, собранных агентом на основе шаблонов
          </p>
        </CardHeader>
      </Card>

      <Card className="flex-1 rounded-4xl shadow-none ">
        <CardBody className="gap-4 px-5 py-5">
          {contracts.length === 0 ? (
            <div className="flex flex-col h-full items-center justify-center py-12 gap-2">
              <DocumentTextIcon className="h-16 w-16 text-default-300" />
              <p className="text-default-400 text-center">Нет сгенерированных документов</p>
              <p className="text-sm text-default-300 text-center">
                Документы будут отображаться здесь после генерации через агента
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {contracts.map((contract) => (
                <div
                  key={contract.id}
                  className="flex flex-col gap-3 p-4 rounded-3xl  bg-default-50/60 hover:bg-default-100/70 transition-colors"
                >
                  <div className="flex items-start gap-3 flex-1">
                    <DocumentTextIcon className="h-6 w-6 flex-shrink-0 text-success mt-1" />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium line-clamp-2">{contract.name}</p>
                      {contract.createdAt && (
                        <p className="text-xs text-default-400 mt-1">
                          {new Date(contract.createdAt).toLocaleString("ru-RU")}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2 mt-3">
                    <Button
                      size="sm"
                      variant="flat"
                      color="secondary"
                      fullWidth
                      startContent={<ArrowDownTrayIcon className="h-4 w-4" />}
                      onPress={() => handleDownload(contract)}
                    >
                      Скачать
                    </Button>
                    <Button
                      isIconOnly
                      size="sm"
                      variant="flat"
                      color="danger"
                      isLoading={contract.deleting}
                      onPress={() => openDeleteModal(contract)}
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardBody>
      </Card>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalContent>
          <ModalHeader>Удалить документ</ModalHeader>
          <ModalBody>
            <p>Вы уверены, что хотите удалить документ &quot;{selectedContract?.name}&quot;?</p>
            <p className="text-sm text-default-400">Это действие нельзя отменить.</p>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onClose}>
              Отмена
            </Button>
            <Button
              color="danger"
              onPress={() => selectedContract && handleDelete(selectedContract)}
            >
              Удалить
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}
