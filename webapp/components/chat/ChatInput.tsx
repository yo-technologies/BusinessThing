"use client";

import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { Button } from "@heroui/button";
import { Textarea } from "@heroui/input";
import { Spinner } from "@heroui/spinner";

interface ChatInputProps {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  disabled?: boolean;
  isStreaming?: boolean;
}

export function ChatInput({
  value,
  onChange,
  onSend,
  disabled,
  isStreaming,
}: ChatInputProps) {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  return (
    <div className="flex items-center gap-2 justify-between p-2 absolute bottom-0 left-0 right-0 bg-gradient-to-t from-content1 to-transparent]">
      <Textarea
        classNames={{
          inputWrapper:
            "border-white/10 border-1 bg-content1/20 backdrop-blur-xs rounded-4xl",
          input: "scrollbar-hide",
        }}
        isDisabled={disabled}
        maxRows={2}
        minRows={1}
        placeholder="Напиши сообщение..."
        radius="full"
        size="md"
        value={value}
        variant="bordered"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
      />
      <Button
        isIconOnly
        color="success"
        isDisabled={disabled || !value.trim()}
        radius="full"
        onPress={onSend}
      >
        {isStreaming ? (
          <Spinner
            classNames={{ wrapper: "w-4 h-4" }}
            color="current"
            size="sm"
          />
        ) : (
          <ArrowRightIcon className="h-5 w-5" />
        )}
      </Button>
    </div>
  );
}
