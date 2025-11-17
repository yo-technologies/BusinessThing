"use client";

import { useState } from "react";
import { Button } from "@heroui/button";

import { useApiClients } from "@/api/client";

interface GeneratedContractResult {
  contract_id: string;
  name: string;
  s3_key: string;
  created_at: string;
  template_name: string;
}

interface ToolCallResultProps {
  toolName: string;
  result: string;
}

/**
 * Компонент для отображения результатов выполнения инструмента
 */
export function ToolCallResult({ toolName, result }: ToolCallResultProps) {
  // Для generate_contract показываем кнопку скачивания
  if (toolName === "generate_contract") {
    return <GeneratedContractResultView result={result} />;
  }

  // Для других инструментов пока ничего не показываем
  return null;
}

function GeneratedContractResultView({ result }: { result: string }) {
  const { core } = useApiClients();
  const [downloading, setDownloading] = useState(false);

  let contractData: GeneratedContractResult | null = null;

  try {
    contractData = JSON.parse(result) as GeneratedContractResult;
  } catch (e) {
    console.error("Failed to parse contract result:", e);

    return null;
  }

  const handleDownload = async () => {
    if (!contractData) return;

    setDownloading(true);
    try {
      // Получаем presigned URL
      const response = await core.v1.storageServiceGenerateDownloadUrl({
        s3Key: contractData.s3_key,
      });

      if (response.data.downloadUrl) {
        // Открываем в новом окне для скачивания
        window.open(response.data.downloadUrl, "_blank");
      }
    } catch (error) {
      console.error("Failed to download contract:", error);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div className="mt-2 flex flex-col gap-2 rounded-lg border border-success-200 bg-success-50/50 p-3">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1">
          <div className="text-sm font-medium text-default-900">
            {contractData.name}
          </div>
          {contractData.template_name && (
            <div className="text-xs text-default-500">
              Шаблон: {contractData.template_name}
            </div>
          )}
        </div>
        <Button
          color="success"
          isLoading={downloading}
          size="sm"
          variant="flat"
          onPress={handleDownload}
        >
          📥 Скачать
        </Button>
      </div>
    </div>
  );
}
