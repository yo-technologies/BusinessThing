"use client";

import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { Switch } from "@heroui/switch"; // Assuming HeroUI Switch component
import { Select, SelectItem } from "@heroui/select"; // Assuming HeroUI Select component

export const DebugPanel = () => {
  const [mounted, setMounted] = useState(false);
  const { theme, setTheme } = useTheme();
  const [role, setRole] = useState("employee"); // Default role

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;

  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-default-100 p-2 flex items-center justify-between text-xs border-b border-default-200">
      <div className="flex items-center gap-2">
        <span>Theme:</span>
        <Switch
          isSelected={theme === "dark"}
          onValueChange={(checked) => setTheme(checked ? "dark" : "light")}
          size="sm"
        >
          {theme === "dark" ? "Dark" : "Light"}
        </Switch>
      </div>
      <div className="flex items-center gap-2">
        <span>Role:</span>
        <Select
          selectedKeys={[role]}
          onSelectionChange={(keys) => setRole(Array.from(keys).join(""))}
          size="sm"
          aria-label="Select Role"
          className="w-[120px]"
        >
          <SelectItem key="admin">
            Admin
          </SelectItem>
          <SelectItem key="employee">
            Employee
          </SelectItem>
        </Select>
      </div>
    </div>
  );
};
