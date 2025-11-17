"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Card } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Button } from "@heroui/button";
import {
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
  useDisclosure,
} from "@heroui/modal";
import { Chip } from "@heroui/chip";
import {
  BuildingOfficeIcon,
  ArrowLeftOnRectangleIcon,
  ShieldCheckIcon,
  UserIcon,
  ArrowUpRightIcon,
  PlusIcon,
} from "@heroicons/react/24/outline";

import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";
import { CoreOrganization } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { Organization } from "@/utils/jwt";

// Combined type for holding both basic and detailed org info
type DetailedOrganization = Organization & { details?: CoreOrganization };

export default function OrganizationsListPage() {
  const router = useRouter();
  const {
    loading: authLoading,
    isAuthenticated,
    isNewUser,
    organizations,
  } = useAuth();
  const { switchOrganization } = useOrganization({
    organizations,
    authLoading,
  });
  const { core } = useApiClients();

  useBackButton(true);

  const { isOpen, onOpen, onClose } = useDisclosure();
  const [selectedOrg, setSelectedOrg] = useState<DetailedOrganization | null>(
    null,
  );
  const [detailedOrgs, setDetailedOrgs] = useState<DetailedOrganization[]>([]);
  const [listLoading, setListLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, authLoading, router]);

  // Fetch details for all organizations on load
  useEffect(() => {
    if (authLoading || organizations.length === 0) {
      setListLoading(false);

      return;
    }

    const fetchAllDetails = async () => {
      setListLoading(true);
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
        console.error("Failed to fetch organization details", error);
      } finally {
        setListLoading(false);
      }
    };

    void fetchAllDetails();
  }, [authLoading, organizations, core.v1]);

  const handleCardPress = (org: DetailedOrganization) => {
    setSelectedOrg(org);
    onOpen();
  };

  const handleOpenOrganization = () => {
    if (!selectedOrg) return;
    switchOrganization(selectedOrg.id);
    router.push("/organization");
  };

  const getRoleLabel = (role?: string) => {
    switch (role) {
      case "admin":
        return {
          label: "Администратор",
          color: "secondary" as const,
          icon: ShieldCheckIcon,
        };
      case "employee":
        return {
          label: "Сотрудник",
          color: "default" as const,
          icon: UserIcon,
        };
      default:
        return {
          label: "Неизвестно",
          color: "default" as const,
          icon: UserIcon,
        };
    }
  };

  if (authLoading || listLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner size="lg" />
      </div>
    );
  }

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Мои организации</h1>
      <div className="mb-4">
        <Button
          color="success"
          variant="flat"
          fullWidth
          startContent={<PlusIcon className="h-5 w-5" />}
          onPress={() => router.push('/organization/create')}
        >
          Создать новую организацию
        </Button>
      </div>
      <div className="flex flex-col gap-3">
        {detailedOrgs.map((org) => {
          const role = getRoleLabel(org.role);
          const RoleIcon = role.icon;

          return (
            <Card
              key={org.id}
              isPressable
              className="flex flex-col gap-3 p-4 rounded-3xl bg-default-50/60 hover:bg-default-100/70 transition-colors"
              onPress={() => handleCardPress(org)}
            >
              <div className="flex items-center gap-3">
                <div className="flex items-center justify-center w-12 h-12 rounded-full bg-default-100 shrink-0">
                  <BuildingOfficeIcon className="h-6 w-6 text-default-400" />
                </div>
                <div className="flex-1 min-w-0 text-left">
                  <p className="font-semibold truncate">
                    {org.details?.name || org.id}
                  </p>
                  <div className="flex gap-2 mt-1">
                    <Chip
                      color={role.color}
                      size="sm"
                      startContent={<RoleIcon className="h-4 w-4" />}
                      variant="flat"
                    >
                      {role.label}
                    </Chip>
                    {org.details?.createdAt && (
                      <Chip color="default" size="sm" variant="flat">
                        Создана:{" "}
                        {new Date(org.details.createdAt).toLocaleDateString(
                          "ru-RU",
                        )}
                      </Chip>
                    )}
                  </div>
                </div>
              </div>
            </Card>
          );
        })}
      </div>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalContent>
          <ModalHeader>
            {selectedOrg?.details?.name || "Организация"}
          </ModalHeader>
          <ModalBody>
            <div>
              <p className="text-sm text-default-500">
                {selectedOrg?.details?.description || "Нет описания"}
              </p>
              {(selectedOrg?.details?.industry ||
                selectedOrg?.details?.region) && (
                <p className="text-xs text-default-400 mt-2">
                  {selectedOrg?.details?.region}
                  {selectedOrg?.details?.industry &&
                    selectedOrg?.details?.region &&
                    " • "}
                  {selectedOrg?.details?.industry}
                </p>
              )}
              {selectedOrg?.details?.createdAt && (
                <p className="text-xs text-default-400 mt-1">
                  Создана:{" "}
                  {new Date(selectedOrg.details.createdAt).toLocaleDateString(
                    "ru-RU",
                  )}
                </p>
              )}
              <div className="mt-6 flex flex-col gap-2">
                <Button
                  color="success"
                  startContent={<ArrowUpRightIcon className="h-5 w-5" />}
                  variant="flat"
                  onPress={handleOpenOrganization}
                >
                  Открыть в приложении
                </Button>
                <Button
                  isDisabled
                  color="danger"
                  startContent={
                    <ArrowLeftOnRectangleIcon className="h-5 w-5" />
                  }
                  variant="flat"
                >
                  Выйти из организации
                </Button>
                {selectedOrg?.role == "admin" && (
                  <p className="text-xs text-default-400 mt-1 text-center w-full">
                    Администратор не может покинуть организацию
                  </p>
                )}
              </div>
            </div>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onClose}>
              Закрыть
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}
