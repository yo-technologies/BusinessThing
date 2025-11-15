"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Chip } from "@heroui/chip";
import { Modal, ModalBody, ModalContent, ModalFooter, ModalHeader, useDisclosure } from "@heroui/modal";
import {
  ArrowUpTrayIcon,
  CheckCircleIcon,
  ClockIcon,
  DocumentIcon,
  ExclamationCircleIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { CoreDocument, CoreDocumentStatus } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";

type DocumentWithLoading = CoreDocument & { deleting?: boolean };

export default function DocumentsPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations });
  const { core } = useApiClients();

  const [documents, setDocuments] = useState<DocumentWithLoading[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDoc, setSelectedDoc] = useState<CoreDocument | null>(null);
  const { isOpen, onOpen, onClose } = useDisclosure();

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

  const loadDocuments = useCallback(async () => {
    if (!currentOrg?.id) return;

    setInitialLoading(true);
    setError(null);
    try {
      const response = await core.v1.documentServiceListDocuments(currentOrg.id, {
        pageSize: 100,
      });
      setDocuments(response.data.documents ?? []);
    } catch (e) {
      console.error("Failed to load documents", e);
      setError("Не удалось загрузить документы");
    } finally {
      setInitialLoading(false);
    }
  }, [core.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadDocuments();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadDocuments]);

  const handleDelete = async (doc: CoreDocument) => {
    if (!doc.id) return;

    setDocuments((prev) =>
      prev.map((d) => (d.id === doc.id ? { ...d, deleting: true } : d)),
    );

    try {
      await core.v1.documentServiceDeleteDocument(doc.id);
      setDocuments((prev) => prev.filter((d) => d.id !== doc.id));
      onClose();
    } catch (e) {
      console.error("Failed to delete document", e);
      setDocuments((prev) =>
        prev.map((d) => (d.id === doc.id ? { ...d, deleting: false } : d)),
      );
    }
  };

  const openDeleteModal = (doc: CoreDocument) => {
    setSelectedDoc(doc);
    onOpen();
  };

  const getStatusLabel = (status?: CoreDocumentStatus) => {
    switch (status) {
      case CoreDocumentStatus.DOCUMENT_STATUS_INDEXED:
        return "Готов";
      case CoreDocumentStatus.DOCUMENT_STATUS_PROCESSING:
        return "Обрабатывается";
      case CoreDocumentStatus.DOCUMENT_STATUS_PENDING:
        return "Ожидает";
      case CoreDocumentStatus.DOCUMENT_STATUS_FAILED:
        return "Ошибка";
      default:
        return "Неизвестно";
    }
  };

  const getStatusColor = (status?: CoreDocumentStatus) => {
    switch (status) {
      case CoreDocumentStatus.DOCUMENT_STATUS_INDEXED:
        return "success";
      case CoreDocumentStatus.DOCUMENT_STATUS_PROCESSING:
        return "warning";
      case CoreDocumentStatus.DOCUMENT_STATUS_PENDING:
        return "default";
      case CoreDocumentStatus.DOCUMENT_STATUS_FAILED:
        return "danger";
      default:
        return "default";
    }
  };

  const getStatusIcon = (status?: CoreDocumentStatus) => {
    switch (status) {
      case CoreDocumentStatus.DOCUMENT_STATUS_INDEXED:
        return <CheckCircleIcon className="h-4 w-4" />;
      case CoreDocumentStatus.DOCUMENT_STATUS_PROCESSING:
        return <ClockIcon className="h-4 w-4" />;
      case CoreDocumentStatus.DOCUMENT_STATUS_PENDING:
        return <ClockIcon className="h-4 w-4" />;
      case CoreDocumentStatus.DOCUMENT_STATUS_FAILED:
        return <ExclamationCircleIcon className="h-4 w-4" />;
      default:
        return <DocumentIcon className="h-4 w-4" />;
    }
  };

  if (authLoading || orgLoading || initialLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Загружаем базу знаний..." color="primary" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-3">
        <h1 className="text-xl font-semibold">Не удалось авторизоваться</h1>
        <p className="text-center text-small text-default-500">
          Попробуй закрыть мини-приложение и открыть его заново из Telegram.
        </p>
      </div>
    );
  }

  const hasDocuments = documents.length > 0;

  return (
    <div className="flex h-full flex-col gap-4">
      <Card className="border-none bg-default-50/60 shadow-sm">
        <CardHeader className="flex items-center justify-between gap-2 pb-2">
          <div className="flex flex-col">
            <span className="text-tiny font-medium uppercase text-primary">
              База знаний
            </span>
            <h1 className="text-xl font-semibold">Документы компании</h1>
          </div>
          {hasDocuments && (
            <div className="flex flex-col items-end gap-1 text-right">
              <span className="text-[11px] text-default-400">
                Всего: {documents.length}
              </span>
              <span className="text-[11px] text-success-500">
                Готово:{" "}
                {documents.filter((d) => d.status === CoreDocumentStatus.DOCUMENT_STATUS_INDEXED).length}
              </span>
            </div>
          )}
        </CardHeader>
        <CardBody className="space-y-2 pb-4">
          <p className="text-small text-default-500">
            Загружай важные документы — договора, отчёты, презентации. Агент
            проиндексирует их и использует в ответах.
          </p>
          {error && <p className="text-xs text-danger-500">{error}</p>}
        </CardBody>
      </Card>

      {hasDocuments ? (
        <div className="grid flex-1 grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {documents.map((doc) => (
            <Card
              key={doc.id}
              className="group relative overflow-hidden border-none bg-content1/80 shadow-sm transition-all hover:shadow-md"
            >
              <CardHeader className="flex items-start justify-between gap-2 pb-2">
                <div className="flex items-start gap-2">
                  <div className="rounded-lg bg-default-100 p-2">
                    <DocumentIcon className="h-5 w-5 text-default-600" />
                  </div>
                  <div className="flex flex-1 flex-col gap-1">
                    <h4 className="line-clamp-2 text-small font-semibold">
                      {doc.name}
                    </h4>
                    <Chip
                      size="sm"
                      variant="flat"
                      color={getStatusColor(doc.status)}
                      startContent={getStatusIcon(doc.status)}
                    >
                      {getStatusLabel(doc.status)}
                    </Chip>
                  </div>
                </div>
                <Button
                  isIconOnly
                  size="sm"
                  variant="light"
                  color="danger"
                  className="opacity-0 transition-opacity group-hover:opacity-100"
                  onPress={() => openDeleteModal(doc)}
                  isDisabled={doc.deleting}
                >
                  <TrashIcon className="h-4 w-4" />
                </Button>
              </CardHeader>
              <CardBody className="pt-0">
                <div className="flex items-center justify-between text-[11px] text-default-400">
                  <span>
                    {doc.fileType?.toUpperCase() ?? "FILE"}
                  </span>
                  <span>
                    {doc.createdAt
                      ? new Date(doc.createdAt).toLocaleDateString("ru-RU", {
                          day: "2-digit",
                          month: "short",
                        })
                      : "—"}
                  </span>
                </div>
                {doc.errorMessage && (
                  <p className="mt-2 text-xs text-danger-500">
                    {doc.errorMessage}
                  </p>
                )}
              </CardBody>
            </Card>
          ))}
        </div>
      ) : (
        <Card className="flex flex-1 items-center justify-center border-none bg-content1/70 shadow-sm">
          <CardBody className="flex flex-col items-center gap-3 text-center">
            <div className="rounded-full bg-default-100 p-4">
              <DocumentIcon className="h-10 w-10 text-default-400" />
            </div>
            <div className="space-y-1">
              <p className="text-medium font-medium">Здесь пока нет документов</p>
              <p className="text-small text-default-500">
                Загрузи первый файл, чтобы начать формировать базу знаний.
              </p>
            </div>
          </CardBody>
        </Card>
      )}

      <div className="fixed bottom-20 right-4 z-40">
        <Button
          isIconOnly
          color="primary"
          size="lg"
          radius="full"
          className="h-14 w-14 shadow-lg shadow-primary/40"
          onPress={() => {
            // TODO: implement file upload
            alert("Загрузка файлов будет реализована позже");
          }}
        >
          <ArrowUpTrayIcon className="h-6 w-6" />
        </Button>
      </div>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalContent>
          <ModalHeader>Удалить документ?</ModalHeader>
          <ModalBody>
            <p className="text-small text-default-500">
              Вы уверены, что хотите удалить <b>{selectedDoc?.name}</b>? Это
              действие нельзя будет отменить.
            </p>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onClose}>
              Отмена
            </Button>
            <Button
              color="danger"
              onPress={() => selectedDoc && handleDelete(selectedDoc)}
              isLoading={(selectedDoc as DocumentWithLoading)?.deleting}
            >
              Удалить
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}