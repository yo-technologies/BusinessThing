"use client";

import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { Switch } from "@heroui/switch"; // Assuming HeroUI Switch component
import { Select, SelectItem } from "@heroui/select"; // Assuming HeroUI Select component
import { useAuth } from "@/hooks/useAuth"; // Import useAuth hook

export const DebugPanel = () => {
  const [mounted, setMounted] = useState(false);
  const { theme, setTheme } = useTheme();
  const { user, loading } = useAuth(); // Use the useAuth hook

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted || loading) return null; // Don't render until mounted and auth state is loaded

  const currentRole = user?.role || 'Employee'; // Get current role from useAuth

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
          selectedKeys={new Set([currentRole.toLowerCase()])} // Use a Set for selectedKeys
          //onSelectionChange={(keys) => {console.log(Array.from(keys).join(""));setRole(Array.from(keys).join("") === 'admin' ? 'Admin' : 'Employee')}} // Use setRole from useAuth
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
