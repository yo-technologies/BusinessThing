"use client";

import { useCallback, useEffect, useState, useRef } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Chip } from "@heroui/chip";
import {
  DocumentIcon,
  TrashIcon,
  CheckCircleIcon,
  ClockIcon,
  ExclamationCircleIcon,
  PlusIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { CoreDocument, CoreDocumentStatus } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";

type DocumentWithLoading = CoreDocument & { deleting?: boolean; uploading?: boolean };

export default function KnowledgePage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations, authLoading });
  const { core } = useApiClients();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [documents, setDocuments] = useState<DocumentWithLoading[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
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

  const loadData = useCallback(async () => {
    if (!currentOrg?.id) return;

    setInitialLoading(true);
    setError(null);
    try {
      const response = await core.v1.documentServiceListDocuments(currentOrg.id, { pageSize: 100 });
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
    void loadData();
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadData]);

  const handleDeleteDocument = async (doc: CoreDocument) => {
    if (!doc.id) return;

    setDocuments((prev) =>
      prev.map((d) => (d.id === doc.id ? { ...d, deleting: true } : d)),
    );

    try {
      await core.v1.documentServiceDeleteDocument(doc.id);
      setDocuments((prev) => prev.filter((d) => d.id !== doc.id));
    } catch (e) {
      console.error("Failed to delete document", e);
      setDocuments((prev) =>
        prev.map((d) => (d.id === doc.id ? { ...d, deleting: false } : d)),
      );
    }
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !currentOrg?.id) return;

    setUploading(true);
    setError(null);

    try {
      // 1. Получить presigned URL
      const uploadUrlResponse = await core.v1.storageServiceGenerateUploadUrl({
        organizationId: currentOrg.id,
        fileName: file.name,
        contentType: file.type || "application/octet-stream",
      });

      const { uploadUrl, s3Key } = uploadUrlResponse.data;

      if (!uploadUrl || !s3Key) {
        throw new Error("Invalid upload URL response");
      }

      // 2. Загрузить файл в S3
      const uploadResponse = await fetch(uploadUrl, {
        method: "PUT",
        body: file,
        headers: {
          "Content-Type": file.type || "application/octet-stream",
        },
      });

      if (!uploadResponse.ok) {
        throw new Error("Failed to upload file to S3");
      }

      // 3. Зарегистрировать документ в базе
      const registerResponse = await core.v1.documentServiceRegisterDocument(currentOrg.id, {
        name: file.name,
        s3Key,
        fileType: file.type || "application/octet-stream",
        fileSize: file.size.toString(),
      });

      // 4. Добавить документ в список
      const newDocument = registerResponse.data.document;
      if (newDocument) {
        setDocuments((prev) => [newDocument, ...prev]);
      }
    } catch (e) {
      console.error("Failed to upload document", e);
      setError("Не удалось загрузить документ");
    } finally {
      setUploading(false);
      // Очистить input для возможности загрузить тот же файл снова
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const getStatusLabel = (status?: CoreDocumentStatus) => {
    switch (status) {
      case CoreDocumentStatus.DOCUMENT_STATUS_INDEXED:
        return { label: "Готов", color: "success" as const, icon: CheckCircleIcon };
      case CoreDocumentStatus.DOCUMENT_STATUS_PROCESSING:
        return { label: "Обработка", color: "warning" as const, icon: ClockIcon };
      case CoreDocumentStatus.DOCUMENT_STATUS_FAILED:
        return { label: "Ошибка", color: "danger" as const, icon: ExclamationCircleIcon };
      default:
        return { label: "В очереди", color: "default" as const, icon: ClockIcon };
    }
  };

  if (authLoading || orgLoading || initialLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error && !documents.length) {
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
      <input
        ref={fileInputRef}
        type="file"
        className="hidden"
        onChange={handleFileSelect}
        accept=".pdf,.doc,.docx,.txt,.md"
      />

      <Card className="rounded-4xl shadow-none">
        <CardHeader className="flex flex-col gap-1 px-5 py-4">
          <div className="flex items-center gap-2 w-full">
            <BookOpenIcon className="h-6 w-6 flex-shrink-0 text-primary" />
            <p className="text-xl font-semibold">База знаний</p>
          </div>
          <p className="text-xs text-default-300">
            Документы, помогающие ассистенту точнее отвечать на вопросы
          </p>
        </CardHeader>
      </Card>

      {error && (
        <Card className="rounded-4xl shadow-sm border border-danger-200/60">
          <CardBody>
            <p className="text-danger text-center">{error}</p>
          </CardBody>
        </Card>
      )}

      <Card className="flex-1 rounded-4xl shadow-none ">
        <CardBody className="gap-4">
          {documents.length === 0 ? (
            <div className="flex flex-col h-full items-center justify-center py-12 gap-2">
              <DocumentIcon className="h-16 w-16 text-default-300" />
              <p className="text-default-400 text-center">Документы отсутствуют</p>
              <p className="text-sm text-default-300 text-center max-w-md">
                Загрузите документы, чтобы агент мог использовать их в работе
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {documents.map((doc) => {
                const status = getStatusLabel(doc.status);
                const StatusIcon = status.icon;

                return (
                  <Card
                    key={doc.id}
                    className="flex flex-row items-center gap-3 p-4 rounded-3xl  bg-default-50/60 hover:bg-default-100/70 transition-colors"
                  >
                    <DocumentIcon className="h-6 w-6 flex-shrink-0 text-primary" />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{doc.name}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <p className="text-sm text-default-400">
                          {doc.fileSize ? `${(Number(doc.fileSize) / 1024).toFixed(1)} KB` : ""}
                        </p>
                        <Chip
                          color={status.color}
                          variant="flat"
                          size="sm"
                          startContent={<StatusIcon className="h-4 w-4" />}
                        >
                          {status.label}
                        </Chip>
                      </div>
                    </div>
                    <Button
                      isIconOnly
                      size="sm"
                      variant="light"
                      color="danger"
                      isLoading={doc.deleting}
                      onPress={() => handleDeleteDocument(doc)}
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </Card>
                );
              })}
            </div>
          )}
        </CardBody>
      </Card>

      <Button
        isIconOnly
        color="secondary"
        size="lg"
        radius="full"
        className="fixed right-6 bottom-24 z-50 shadow-lg"
        onPress={handleUploadClick}
        isLoading={uploading}
        isDisabled={uploading}
      >
        <PlusIcon className="h-6 w-6" />
      </Button>
    </div>
  );
}
