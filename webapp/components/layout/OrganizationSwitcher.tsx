"use client";

import { Button } from "@heroui/button";
import { Dropdown, DropdownItem, DropdownMenu, DropdownTrigger } from "@heroui/dropdown";
import { ChevronDownIcon, BuildingOfficeIcon } from "@heroicons/react/24/outline";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";

export const OrganizationSwitcher = () => {
  const { organizations, loading } = useAuth();
  const { currentOrg, switchOrganization } = useOrganization({ organizations, authLoading: loading });

  if (!currentOrg || organizations.length <= 1) {
    return null;
  }

  return (
    <Dropdown placement="bottom-start">
      <DropdownTrigger>
        <Button
          variant="flat"
          size="sm"
          className="gap-2"
          endContent={<ChevronDownIcon className="h-4 w-4" />}
          startContent={<BuildingOfficeIcon className="h-4 w-4" />}
        >
          <span className="max-w-[150px] truncate">{currentOrg.id}</span>
        </Button>
      </DropdownTrigger>
      <DropdownMenu
        aria-label="Переключение организации"
        selectedKeys={currentOrg.id ? [currentOrg.id] : []}
        selectionMode="single"
        onSelectionChange={(keys) => {
          const key = Array.from(keys)[0] as string;
          if (key) {
            switchOrganization(key);
          }
        }}
      >
        {organizations.map((org) => (
          <DropdownItem key={org.id} textValue={org.id}>
            <div className="flex flex-col">
              <span className="font-medium">{org.id}</span>
              <span className="text-xs text-default-400">{org.role}</span>
            </div>
          </DropdownItem>
        ))}
      </DropdownMenu>
    </Dropdown>
  );
};
