"use client";

import { useEffect, useState } from "react";
import { Button } from "@heroui/button";
import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownTrigger,
} from "@heroui/dropdown";
import {
  ChevronDownIcon,
  BuildingOfficeIcon,
} from "@heroicons/react/24/outline";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useApiClients } from "@/api/client";
import { CoreOrganization } from "@/api/api.core.generated";
import { Organization } from "@/utils/jwt";

// Define the detailed type
type DetailedOrganization = Organization & { details?: CoreOrganization };

export const OrganizationSwitcher = () => {
  const { organizations, loading: authLoading } = useAuth();
  const { currentOrg, switchOrganization } = useOrganization({
    organizations,
    authLoading,
  });
  const { core } = useApiClients();

  const [detailedOrgs, setDetailedOrgs] = useState<DetailedOrganization[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Effect to fetch details
  useEffect(() => {
    if (authLoading || organizations.length === 0) {
      setIsLoading(false);

      return;
    }

    const fetchAllDetails = async () => {
      setIsLoading(true);
      try {
        const orgDetailsPromises = organizations.map((org) =>
          core.v1.organizationServiceGetOrganization(org.id),
        );
        const responses = await Promise.all(orgDetailsPromises);
        const orgsWithDetails = organizations.map((org, index) => ({
          ...org,
          details: responses[index].data.organization || undefined,
        }));

        setDetailedOrgs(orgsWithDetails);
      } catch (error) {
        console.error(
          "Failed to fetch organization details for switcher",
          error,
        );
      } finally {
        setIsLoading(false);
      }
    };

    void fetchAllDetails();
  }, [authLoading, organizations, core.v1]);

  // Find the detailed info for the current org
  const currentDetailedOrg = detailedOrgs.find(
    (org) => org.id === currentOrg?.id,
  );
  const orgName = currentDetailedOrg?.details?.name || currentOrg?.id;

  if (!currentOrg) {
    return null;
  }

  // If only one org, just display the name without a dropdown
  if (organizations.length <= 1) {
    return (
      <Button
        className="gap-2 backdrop-blur-xs"
        size="sm"
        startContent={<BuildingOfficeIcon className="h-4 w-4" />}
        variant="flat"
      >
        <span className="max-w-[150px] truncate">{orgName}</span>
      </Button>
    );
  }

  return (
    <Dropdown>
      <DropdownTrigger>
        <Button
          radius="full"
          className="gap-2 backdrop-blur-xs"
          endContent={<ChevronDownIcon className="h-4 w-4" />}
          size="sm"
          startContent={<BuildingOfficeIcon className="h-4 w-4" />}
          variant="flat"
        >
          <span className="max-w-[150px] truncate">{orgName}</span>
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
        // isLoading={isLoading}
      >
        {detailedOrgs.map((org) => (
          <DropdownItem key={org.id} textValue={org.details?.name || org.id}>
            <div className="flex flex-col">
              <span className="font-medium">{org.details?.name || org.id}</span>
              <span className="text-xs text-default-400">{org.role}</span>
            </div>
          </DropdownItem>
        ))}
      </DropdownMenu>
    </Dropdown>
  );
};
