"use client";

import { AgentToolCallStatus } from "@/api/api.agent.generated";

interface ToolCallStatusProps {
  status: AgentToolCallStatus;
}

/**
 * Компонент для отображения статуса вызова инструмента
 * Компактная версия - только иконка
 */
export function ToolCallStatus({ status }: ToolCallStatusProps) {
  const statusConfig = getStatusConfig(status);

  return (
    <span className="text-xs" title={statusConfig.label}>
      {statusConfig.icon}
    </span>
  );
}

interface StatusConfig {
  label: string;
  color: "default" | "primary" | "success" | "warning" | "danger";
  icon: string;
}

function getStatusConfig(status: AgentToolCallStatus): StatusConfig {
  switch (status) {
    case AgentToolCallStatus.TOOL_CALL_STATUS_PENDING:
      return {
        label: "Ожидает",
        color: "default",
        icon: "⏳",
      };
    case AgentToolCallStatus.TOOL_CALL_STATUS_EXECUTING:
      return {
        label: "Выполняется",
        color: "primary",
        icon: "⚙️",
      };
    case AgentToolCallStatus.TOOL_CALL_STATUS_COMPLETED:
      return {
        label: "Выполнено",
        color: "success",
        icon: "✅",
      };
    case AgentToolCallStatus.TOOL_CALL_STATUS_FAILED:
      return {
        label: "Ошибка",
        color: "danger",
        icon: "❌",
      };
    default:
      return {
        label: "Неизвестно",
        color: "default",
        icon: "❓",
      };
  }
}
