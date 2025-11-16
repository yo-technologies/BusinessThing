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

// Словарь перевода имен агентов
const AGENT_NAMES: Record<string, string> = {
  "main": "Андрей",
  "marketing_agent": "Маркетолог Лаврентий",
  "legal_agent": "Юрист Юра",
  "business_analyst_agent": "Бизнес-аналитик Игорь",
};

// Функция для получения имени агента
const getAgentName = (sender?: string): string => {
  if (!sender) return "Ассистент";
  return AGENT_NAMES[sender] || "Ассистент";
};

export function ChatWindow({ messages, streamingMessage, streamingToolCalls, isStreaming, loadingMessages }: ChatWindowProps) {
  // Конвертируем streaming tool calls в формат AgentToolCall для единообразного отображения
  const convertStreamingToolCall = (event: AgentToolCallEvent): AgentToolCall => {
    return {
      id: event.toolCallId || "",
      name: event.toolName || "",
      arguments: event.arguments || "{}",
      result: undefined,
      status: event.status as AgentToolCallStatus || AgentToolCallStatus.TOOL_CALL_STATUS_PENDING,
      createdAt: new Date().toISOString(),
    };
  };

  return (
    <Card className="flex flex-1 flex-col border-none bg-content1/70 shadow-none">
      <CardBody className="flex flex-1 flex-col gap-3 overflow-y-auto pb-4">
        {loadingMessages ? (
          <div className="flex flex-1 flex-row h-full justify-center items-center">
            <Spinner size="sm" label="Загружаем сообщения..." color="primary" />
          </div>
        ) : messages.length === 0 && !streamingMessage && (!streamingToolCalls || streamingToolCalls.size === 0) ? (
            <div className="flex flex-1 flex-row h-full justify-center items-center">
                <span className="text-center text-small text-default-400 h-fit p-4">
                Пока сообщений нет. Напиши что-нибудь, чтобы начать диалог.
                </span>
            </div>
        ) : (
          <>
            {messages.map((message, index) => {
              // Пропускаем системные сообщения и сообщения типа tool
              if (
                message.role === AgentMessageRole.MESSAGE_ROLE_SYSTEM ||
                message.role === AgentMessageRole.MESSAGE_ROLE_TOOL
              ) {
                return null;
              }

              const isUser = message.role === AgentMessageRole.MESSAGE_ROLE_USER;
              const isAssistant = message.role === AgentMessageRole.MESSAGE_ROLE_ASSISTANT;
              
              // Находим предыдущее отображаемое сообщение (не system и не tool)
              let prevVisibleMessage = null;
              for (let i = index - 1; i >= 0; i--) {
                const prevMsg = messages[i];
                if (
                  prevMsg.role !== AgentMessageRole.MESSAGE_ROLE_SYSTEM &&
                  prevMsg.role !== AgentMessageRole.MESSAGE_ROLE_TOOL
                ) {
                  prevVisibleMessage = prevMsg;
                  break;
                }
              }
              
              // Определяем, нужно ли показывать заголовок для сообщений агента
              const showAssistantHeader = isAssistant && (
                !prevVisibleMessage || 
                prevVisibleMessage.role !== AgentMessageRole.MESSAGE_ROLE_ASSISTANT ||
                prevVisibleMessage.sender !== message.sender
              );
              
              return (
                <div key={message.id} className="flex flex-col gap-2">
                  {/* Заголовок агента (только для первого сообщения в последовательности от этого агента) */}
                  {showAssistantHeader && (
                    <div className="text-small font-bold text-default-600">
                      {getAgentName(message.sender)}
                    </div>
                  )}
                  
                  {/* Отображаем текст сообщения */}
                  {message.content && (
                    <div
                      className={
                        isUser
                          ? "ml-auto max-w-[80%] rounded-xl bg-primary text-primary-foreground px-3 py-2 text-small"
                          : "w-full text-small"
                      }
                    >
                      <MarkdownWrapper content={message.content} />
                    </div>
                  )}
                  
                  {/* Отображаем tool calls сообщения, если есть */}
                  {message.toolCalls && message.toolCalls.length > 0 && (
                    <div className="flex flex-col gap-2">
                      {message.toolCalls.map((toolCall) => (
                        <ToolCallMessage key={toolCall.id} toolCall={toolCall} />
                      ))}
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
                    key={toolCallEvent.toolCallId} 
                    toolCall={convertStreamingToolCall(toolCallEvent)} 
                  />
                ))}
              </div>
            )}
            
            {/* Отображаем streaming message */}
            {streamingMessage && (
              <div className="flex flex-col gap-2">
                {/* Заголовок для streaming сообщения агента, если последнее видимое сообщение не от ассистента */}
                {(() => {
                  // Находим последнее видимое сообщение
                  for (let i = messages.length - 1; i >= 0; i--) {
                    const msg = messages[i];
                    if (
                      msg.role !== AgentMessageRole.MESSAGE_ROLE_SYSTEM &&
                      msg.role !== AgentMessageRole.MESSAGE_ROLE_TOOL
                    ) {
                      // Показываем заголовок, если последнее сообщение не от ассистента
                      return msg.role !== AgentMessageRole.MESSAGE_ROLE_ASSISTANT ? (
                        <div className="text-small font-semibold text-default-600 px-1">
                          Ассистент
                        </div>
                      ) : null;
                    }
                  }
                  // Если нет видимых сообщений, показываем заголовок
                  return (
                    <div className="text-small font-semibold text-default-600 px-1">
                      Ассистент
                    </div>
                  );
                })()}
                <div className="w-full rounded-xl bg-default-100 px-3 py-2 text-small">
                  <MarkdownWrapper content={streamingMessage} />
                  {isStreaming && <span className="ml-1 animate-pulse">▍</span>}
                </div>
              </div>
            )}
          </>
        )}
      </CardBody>
    </Card>
  );
}
