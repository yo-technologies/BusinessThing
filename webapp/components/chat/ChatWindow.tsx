"use client";

import { Card, CardBody } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { AgentMessage, AgentMessageRole, AgentToolCall, AgentToolCallEvent, AgentToolCallStatus } from "@/api/api.agent.generated";
import { MarkdownWrapper } from "@/components/MarkdownWrapper";
import { ToolCallMessage } from "./ToolCallMessage";

interface ChatWindowProps {
  messages: AgentMessage[];
  streamingMessage?: string;
  streamingToolCalls?: Map<string, AgentToolCallEvent>;
  isStreaming?: boolean;
  loadingMessages?: boolean;
}

export function ChatWindow({ messages, streamingMessage, streamingToolCalls, isStreaming, loadingMessages }: ChatWindowProps) {
  // Конвертируем streaming tool calls в формат AgentToolCall для единообразного отображения
  const convertStreamingToolCall = (event: AgentToolCallEvent): AgentToolCall => {
    return {
      id: event.toolName || "",
      name: event.toolName || "",
      arguments: event.arguments || "{}",
      result: undefined,
      status: event.status as AgentToolCallStatus || AgentToolCallStatus.TOOL_CALL_STATUS_PENDING,
      createdAt: new Date().toISOString(),
    };
  };

  return (
    <Card className="flex flex-1 flex-col border-none bg-content1/70 shadow-sm rounded-2xl">
      <CardBody className="flex flex-1 flex-col gap-3 overflow-y-auto pb-4">
        {loadingMessages ? (
          <div className="flex flex-1 flex-row h-full justify-center items-center">
            <Spinner size="sm" label="Загружаем сообщения..." color="primary" />
          </div>
        ) : messages.length === 0 && !streamingMessage && (!streamingToolCalls || streamingToolCalls.size === 0) ? (
            <div className="flex flex-1 flex-row h-full justify-center items-center">
                <span className="text-center text-small text-default-400 h-fit">
                Пока сообщений нет. Напиши что-нибудь, чтобы начать диалог.
                </span>
            </div>
        ) : (
          <>
            {messages.map((message) => {
              const isUser = message.role === AgentMessageRole.MESSAGE_ROLE_USER;
              
              return (
                <div key={message.id} className="flex flex-col gap-2">
                  {/* Отображаем tool calls сообщения, если есть */}
                  {message.toolCalls && message.toolCalls.length > 0 && (
                    <div className="flex flex-col gap-2">
                      {message.toolCalls.map((toolCall) => (
                        <ToolCallMessage key={toolCall.id} toolCall={toolCall} />
                      ))}
                    </div>
                  )}
                  
                  {/* Отображаем текст сообщения */}
                  {message.content && (
                    <div
                      className={
                        isUser
                          ? "ml-auto max-w-[80%] rounded-xl bg-primary text-primary-foreground px-3 py-2 text-small"
                          : "mr-auto max-w-[80%] rounded-xl bg-default-100 px-3 py-2 text-small"
                      }
                    >
                      <MarkdownWrapper content={message.content} />
                    </div>
                  )}
                </div>
              );
            })}
            
            {/* Отображаем streaming tool calls */}
            {streamingToolCalls && streamingToolCalls.size > 0 && (
              <div className="flex flex-col gap-2">
                {Array.from(streamingToolCalls.values()).map((toolCallEvent) => (
                  <ToolCallMessage 
                    key={toolCallEvent.toolName} 
                    toolCall={convertStreamingToolCall(toolCallEvent)} 
                  />
                ))}
              </div>
            )}
            
            {/* Отображаем streaming message */}
            {streamingMessage && (
              <div className="mr-auto max-w-[80%] rounded-xl bg-default-100 px-3 py-2 text-small">
                <MarkdownWrapper content={streamingMessage} />
                {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
              </div>
            )}
          </>
        )}
      </CardBody>
    </Card>
  );
}
