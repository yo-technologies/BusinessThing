"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@heroui/button";
import {
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownSection,
  DropdownTrigger,
} from "@heroui/dropdown";
import {
  ChevronDownIcon,
  BuildingOfficeIcon,
  PlusIcon,
} from "@heroicons/react/24/outline";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useApiClients } from "@/api/client";
import { CoreOrganization } from "@/api/api.core.generated";
import { Organization } from "@/utils/jwt";

// Define the detailed type
type DetailedOrganization = Organization & { details?: CoreOrganization };

export const OrganizationSwitcher = () => {
  const router = useRouter();
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

  return (
    <Dropdown radius="full" className="h-7">
      <DropdownTrigger>
        <Button
          className="gap-2 backdrop-blur-xs h-7"
          endContent={<ChevronDownIcon className="h-4 w-4" />}
          radius="full"
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

          if (key === "new-organization") {
            router.push("/organization/create");
          } else if (key) {
            switchOrganization(key);
          }
        }}
      >
        <DropdownSection title="Ваши организации" showDivider>
          {detailedOrgs.map((org) => (
            <DropdownItem key={org.id} textValue={org.details?.name || org.id} color="default">
              <div className="flex flex-col">
                <span className="font-medium">
                  {org.details?.name || org.id}
                </span>
                <span className="text-xs text-default-600">{org.role}</span>
              </div>
            </DropdownItem>
          ))}
        </DropdownSection>
        <DropdownSection>
          <DropdownItem
            key="new-organization"
            startContent={<PlusIcon className="h-8 w-5" />}
            className="text-success"
            color="success"
          >
            Создать организацию
          </DropdownItem>
        </DropdownSection>
      </DropdownMenu>
    </Dropdown>
  );
};
