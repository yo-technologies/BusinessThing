"use client";

import { AgentToolCall } from "@/api/api.agent.generated";
import { getToolInfo } from "@/utils/toolNames";
import { ToolCallStatus } from "./ToolCallStatus";

interface ToolCallMessageProps {
  toolCall: AgentToolCall;
}

/**
 * Компактный компонент для отображения вызова инструмента
 */
export function ToolCallMessage({ toolCall }: ToolCallMessageProps) {
  const toolInfo = getToolInfo(toolCall.name ?? "");

  return (
    <div className="mr-auto flex max-w-[70%] items-center gap-1.5 rounded-lg border border-default-200 bg-default-50 px-2 py-1 text-xs">
      <span className="text-sm">{toolInfo.icon}</span>
      <span className="flex-1 font-medium text-default-700">{toolInfo.displayName}</span>
      <ToolCallStatus status={toolCall.status!} />
    </div>
  );
}
