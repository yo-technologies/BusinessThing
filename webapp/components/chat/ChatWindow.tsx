"use client";

import { Card, CardBody } from "@heroui/card";
import { AgentMessage, AgentMessageRole } from "@/api/api.agent.generated";
import { MarkdownWrapper } from "@/components/MarkdownWrapper";

interface ChatWindowProps {
  messages: AgentMessage[];
  streamingMessage?: string;
  isStreaming?: boolean;
}

export function ChatWindow({ messages, streamingMessage, isStreaming }: ChatWindowProps) {
  return (
    <Card className="flex flex-1 flex-col border-none bg-content1/70 shadow-sm rounded-2xl">
      <CardBody className="flex flex-1 flex-col gap-3 overflow-y-auto pb-4">
        {messages.length === 0 && !streamingMessage ? (
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
                <div
                  key={message.id}
                  className={
                    isUser
                      ? "ml-auto max-w-[80%] rounded-xl bg-primary text-primary-foreground px-3 py-2 text-small"
                      : "mr-auto max-w-[80%] rounded-xl bg-default-100 px-3 py-2 text-small"
                  }
                >
                  <MarkdownWrapper content={message.content} />
                </div>
              );
            })}
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
