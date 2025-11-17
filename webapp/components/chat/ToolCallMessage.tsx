"use client";

import { ToolCallStatus } from "./ToolCallStatus";
import { ToolCallResult } from "./ToolCallResult";

import { AgentToolCall, AgentToolCallStatus } from "@/api/api.agent.generated";
import { getToolInfo } from "@/utils/toolNames";

interface ToolCallMessageProps {
  toolCall: AgentToolCall;
}

/**
 * Компактный компонент для отображения вызова инструмента
 */
export function ToolCallMessage({ toolCall }: ToolCallMessageProps) {
  const toolInfo = getToolInfo(toolCall.name ?? "");
  const hasResult = 
    toolCall.status === AgentToolCallStatus.TOOL_CALL_STATUS_COMPLETED &&
    toolCall.result;

  return (
    <div className="mr-auto flex max-w-[70%] flex-col gap-2">
      <div className="flex items-center gap-1.5 rounded-lg border border-success-200/50 bg-default-100 px-2 py-1 text-xs">
        <span className="text-sm">{toolInfo.icon}</span>
        <span className="flex-1 font-medium text-default-700">
          {toolInfo.displayName}
        </span>
        <ToolCallStatus status={toolCall.status!} />
      </div>
      
      {hasResult && (
        <ToolCallResult 
          toolName={toolCall.name ?? ""} 
          result={toolCall.result ?? ""} 
        />
      )}
    </div>
  );
}
