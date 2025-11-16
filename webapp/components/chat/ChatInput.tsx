"use client";

import { ArrowRightIcon } from "@heroicons/react/24/outline";
import { Button } from "@heroui/button";
import { Input } from "@heroui/input";
import { Spinner } from "@heroui/spinner";

interface ChatInputProps {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  disabled?: boolean;
  isStreaming?: boolean;
}

export function ChatInput({ value, onChange, onSend, disabled, isStreaming }: ChatInputProps) {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  return (
    <div className="flex items-center gap-2 justify-between">
      <Input
        size="md"
        radius="full"
        variant="bordered"
        classNames={{inputWrapper: "border-white/10 border-1"}}
        placeholder="Напиши сообщение..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        isDisabled={disabled}
      />
      <Button
        isIconOnly
        color="primary"
        radius="full"
        onPress={onSend}
        isDisabled={disabled || !value.trim()}
      >
        {isStreaming ? <Spinner classNames={{wrapper: "w-3 h-3"}} color="default"/> : <ArrowRightIcon className="h-5 w-5" />}
      </Button>
    </div>
  );
}
