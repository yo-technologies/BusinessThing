"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Chip } from "@heroui/chip";
import {
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
  useDisclosure,
} from "@heroui/modal";
import { Select, SelectItem } from "@heroui/select";
import { Tabs, Tab } from "@heroui/tabs";
import {
  UserIcon,
  ShieldCheckIcon,
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  EnvelopeIcon,
  UsersIcon,
  PlusIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";
import { shareURL } from "@telegram-apps/sdk-react";

import {
  CoreUser,
  CoreUserRole,
  CoreUserStatus,
  CoreInvitation,
} from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useHasInvitation } from "@/hooks/useInvitationToken";
import { useBackButton } from "@/hooks/useBackButton";
import { useCurrentRole } from "@/hooks/useCurrentRole";

type TabType = "users" | "invitations";

export default function UsersPage() {
  const router = useRouter();
  const {
    loading: authLoading,
    isAuthenticated,
    isNewUser,
    organizations,
    userInfo,
  } = useAuth();
  const {
    currentOrg,
    loading: orgLoading,
    needsOrganization,
  } = useOrganization({ organizations, authLoading });
  const hasInvitation = useHasInvitation();
  const { core } = useApiClients();
  const { isAdmin } = useCurrentRole();

  const [activeTab, setActiveTab] = useState<TabType>("users");
  const [users, setUsers] = useState<CoreUser[]>([]);
  const [invitations, setInvitations] = useState<CoreInvitation[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inviteRole, setInviteRole] = useState<CoreUserRole>(
    CoreUserRole.USER_ROLE_EMPLOYEE,
  );
  const [inviting, setInviting] = useState(false);
  const [invitationUrl, setInvitationUrl] = useState<string | null>(null);
  const [deletingInvitations, setDeletingInvitations] = useState<Set<string>>(
    new Set(),
  );

  const {
    isOpen: isInviteModalOpen,
    onOpen: onInviteModalOpen,
    onClose: onInviteModalClose,
  } = useDisclosure();

  useBackButton(true);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, authLoading, router]);

  useEffect(() => {
    if (
      !authLoading &&
      !orgLoading &&
      isAuthenticated &&
      !isNewUser &&
      needsOrganization
    ) {
      if (hasInvitation) {
        router.replace("/invitation");
      } else {
        router.replace("/organization/create");
      }
    }
  }, [
    authLoading,
    orgLoading,
    isAuthenticated,
    isNewUser,
    needsOrganization,
    hasInvitation,
    router,
  ]);

  const loadUsers = useCallback(async () => {
    if (!currentOrg?.id) return;

    setInitialLoading(true);
    setError(null);
    try {
      const response = await core.v1.userServiceListUsers(currentOrg.id, {
        pageSize: 100,
      });

      setUsers(response.data.users ?? []);
    } catch (e) {
      console.error("Failed to load users", e);
      setError("Не удалось загрузить список сотрудников");
    } finally {
      setInitialLoading(false);
    }
  }, [core.v1, currentOrg?.id]);

  const loadInvitations = useCallback(async () => {
    if (!currentOrg?.id) return;

    try {
      const response = await core.v1.userServiceListInvitations(currentOrg.id, {
        pageSize: 100,
      });

      setInvitations(response.data.invitations ?? []);
    } catch (e) {
      console.error("Failed to load invitations", e);
    }
  }, [core.v1, currentOrg?.id]);

  useEffect(() => {
    if (!isAuthenticated || authLoading || isNewUser || !currentOrg?.id) return;
    void loadUsers();
    void loadInvitations();
  }, [
    isAuthenticated,
    authLoading,
    isNewUser,
    currentOrg?.id,
    loadUsers,
    loadInvitations,
  ]);

  const handleInvite = async () => {
    if (!currentOrg?.id) return;

    setInviting(true);
    try {
      const response = await core.v1.userServiceInviteUser(currentOrg.id, {
        role: inviteRole,
      });

      if (response.data.invitationUrl) {
        onInviteModalClose();
        void loadInvitations();

        // Сразу открываем шаринг в Telegram
        try {
          if (shareURL.isAvailable()) {
            shareURL(
              response.data.invitationUrl,
              "Присоединяйтесь к нашей организации в BusinessThing!",
            );
          } else {
            // Если шаринг недоступен, копируем в буфер
            navigator.clipboard.writeText(response.data.invitationUrl);
          }
        } catch (shareError) {
          console.error("Failed to share invitation:", shareError);
          // При ошибке шаринга копируем ссылку
          navigator.clipboard.writeText(response.data.invitationUrl);
        }
      }
    } catch (e) {
      console.error("Failed to invite user", e);
    } finally {
      setInviting(false);
    }
  };

  const handleCopyInvitation = () => {
    if (invitationUrl) {
      navigator.clipboard.writeText(invitationUrl);
    }
  };

  const handleShareInvitation = () => {
    if (!invitationUrl) return;

    try {
      if (shareURL.isAvailable()) {
        shareURL(
          invitationUrl,
          "Присоединяйтесь к нашей организации в BusinessThing!",
        );
      } else {
        // Fallback для браузера
        if (navigator.share) {
          navigator
            .share({
              title: "Приглашение в организацию",
              text: "Присоединяйтесь к нашей организации в BusinessThing!",
              url: invitationUrl,
            })
            .catch((err) => console.error("Share failed:", err));
        } else {
          // Если шаринг недоступен, просто копируем
          handleCopyInvitation();
        }
      }
    } catch (error) {
      console.error("Failed to share invitation:", error);
      handleCopyInvitation();
    }
  };

  const handleDeleteInvitation = async (invitationId: string) => {
    if (!invitationId) return;

    setDeletingInvitations((prev) => new Set(prev).add(invitationId));
    try {
      await core.v1.userServiceDeleteInvitation(invitationId);
      void loadInvitations();
    } catch (e) {
      console.error("Failed to delete invitation", e);
    } finally {
      setDeletingInvitations((prev) => {
        const next = new Set(prev);

        next.delete(invitationId);

        return next;
      });
    }
  };

  const getRoleLabel = (role?: CoreUserRole) => {
    switch (role) {
      case CoreUserRole.USER_ROLE_ADMIN:
        return {
          label: "Администратор",
          color: "secondary" as const,
          icon: ShieldCheckIcon,
        };
      case CoreUserRole.USER_ROLE_EMPLOYEE:
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

  const getStatusLabel = (status?: CoreUserStatus) => {
    switch (status) {
      case CoreUserStatus.USER_STATUS_ACTIVE:
        return {
          label: "Активен",
          color: "success" as const,
          icon: CheckCircleIcon,
        };
      case CoreUserStatus.USER_STATUS_PENDING:
        return { label: "Ожидает", color: "warning" as const, icon: ClockIcon };
      case CoreUserStatus.USER_STATUS_INACTIVE:
        return {
          label: "Неактивен",
          color: "danger" as const,
          icon: XCircleIcon,
        };
      default:
        return {
          label: "Неизвестно",
          color: "default" as const,
          icon: ClockIcon,
        };
    }
  };

  if (authLoading || orgLoading || initialLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <Card className="max-w-xl mx-auto mt-8 rounded-4xl shadow-sm border border-danger-200/60">
        <CardBody>
          <p className="text-danger text-center">{error}</p>
        </CardBody>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4 flex-1">
      <Card className="rounded-4xl shadow-none">
        <CardHeader className="flex flex-col gap-1 px-5 py-4">
          <div className="flex items-center gap-2 w-full">
            <UsersIcon className="h-6 w-6 flex-shrink-0 text-secondary" />
            <p className="text-xl font-semibold">Сотрудники</p>
          </div>
          <p className="text-xs text-default-300">
            Управление пользователями и приглашениями в вашу организацию
          </p>
        </CardHeader>
      </Card>

      <Card className="flex-1 rounded-4xl shadow-none h-full">
        <CardBody className="gap-2">
          {/* Табы */}
          <Tabs
            fullWidth
            aria-label="Управление командой"
            classNames={{
              tabWrapper: "bg-default-100/60 rounded-full p-1",
              tabList: "w-full relative rounded-full p-0",
              cursor: "rounded-full",
              panel: "p-0",
            }}
            selectedKey={activeTab}
            size="lg"
            onSelectionChange={(key) => setActiveTab(key as TabType)}
          >
            <Tab
              key="users"
              className="h-full"
              title={
                <div className="flex items-center gap-2 rounded-full">
                  <UsersIcon className="h-4 w-4" />
                  <span>Пользователи</span>
                  <Chip color="secondary" size="sm">
                    {users.length}
                  </Chip>
                </div>
              }
            >
              {users.length === 0 ? (
                <div className="flex flex-col h-full items-center justify-center py-12 gap-2">
                  <UserIcon className="h-16 w-16 text-default-300" />
                  <p className="text-default-400 text-center">
                    Нет сотрудников
                  </p>
                </div>
              ) : (
                <div className="flex flex-col gap-3">
                  {users.map((user) => {
                    const role = getRoleLabel(user.role);
                    const status = getStatusLabel(user.status);
                    const RoleIcon = role.icon;
                    const StatusIcon = status.icon;

                    return (
                      <Card
                        key={user.id}
                        className="flex flex-col gap-3 p-4 rounded-3xl  bg-default-50/60 hover:bg-default-100/70 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-default-100 shrink-0">
                            <UserIcon className="h-6 w-6 text-default-400" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                              <p className="font-semibold">
                                {user.firstName} {user.lastName}
                              </p>
                              {user.id === userInfo?.userId && (
                                <Chip color="default" size="sm" variant="flat">
                                  Вы
                                </Chip>
                              )}
                            </div>
                            <p className="text-sm text-default-400 truncate">
                              {user.email || "Нет email"}
                            </p>
                          </div>
                        </div>
                        <div className="flex gap-2 flex-wrap">
                          <Chip
                            color={role.color}
                            size="sm"
                            startContent={<RoleIcon className="h-4 w-4" />}
                            variant="flat"
                          >
                            {role.label}
                          </Chip>
                          <Chip
                            color={status.color}
                            size="sm"
                            startContent={<StatusIcon className="h-4 w-4" />}
                            variant="flat"
                          >
                            {status.label}
                          </Chip>
                        </div>
                      </Card>
                    );
                  })}
                </div>
              )}
            </Tab>

            <Tab
              key="invitations"
              className="h-full"
              title={
                <div className="flex items-center gap-2">
                  <EnvelopeIcon className="h-4 w-4" />
                  <span>Приглашения</span>
                  <Chip color="secondary" size="sm">
                    {invitations.length}
                  </Chip>
                </div>
              }
            >
              {invitations.length === 0 ? (
                <div className="flex flex-col h-full items-center justify-center py-12 gap-2">
                  <EnvelopeIcon className="h-16 w-16 text-default-300" />
                  <p className="text-default-400 text-center">
                    Нет активных приглашений
                  </p>
                  <p className="text-xs text-default-300 text-center max-w-xs">
                    Создайте приглашение, чтобы добавить новых участников в
                    команду
                  </p>
                </div>
              ) : (
                <div className="flex flex-col gap-3">
                  {invitations.map((invitation) => {
                    const role = getRoleLabel(invitation.role);
                    const RoleIcon = role.icon;
                    const isExpired = invitation.expiresAt
                      ? new Date(invitation.expiresAt) < new Date()
                      : false;
                    const isUsed = !!invitation.usedAt;

                    return (
                      <div
                        key={invitation.id}
                        className="flex flex-col gap-3 p-4 rounded-3xl  bg-default-50/60 hover:bg-default-100/70 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-default-100 shrink-0">
                            <EnvelopeIcon className="h-6 w-6 text-default-400" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold">
                              Приглашение #{invitation.id?.slice(0, 8)}
                            </p>
                            <p className="text-xs text-default-400">
                              Создано:{" "}
                              {invitation.createdAt
                                ? new Date(
                                    invitation.createdAt,
                                  ).toLocaleDateString("ru-RU")
                                : "N/A"}
                            </p>
                          </div>
                          {isAdmin && !isUsed && (
                            <Button
                              isIconOnly
                              color="danger"
                              isDisabled={deletingInvitations.has(
                                invitation.id!,
                              )}
                              isLoading={deletingInvitations.has(
                                invitation.id!,
                              )}
                              size="sm"
                              variant="light"
                              onPress={() =>
                                handleDeleteInvitation(invitation.id!)
                              }
                            >
                              <TrashIcon className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                        <div className="flex gap-2 flex-wrap">
                          <Chip
                            color={role.color}
                            size="sm"
                            startContent={<RoleIcon className="h-4 w-4" />}
                            variant="flat"
                          >
                            {role.label}
                          </Chip>
                          {isUsed ? (
                            <Chip
                              color="success"
                              size="sm"
                              startContent={
                                <CheckCircleIcon className="h-4 w-4" />
                              }
                              variant="flat"
                            >
                              Использовано
                            </Chip>
                          ) : isExpired ? (
                            <Chip
                              color="danger"
                              size="sm"
                              startContent={<XCircleIcon className="h-4 w-4" />}
                              variant="flat"
                            >
                              Истекло
                            </Chip>
                          ) : (
                            <Chip
                              color="warning"
                              size="sm"
                              startContent={<ClockIcon className="h-4 w-4" />}
                              variant="flat"
                            >
                              Активно
                            </Chip>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </Tab>
          </Tabs>
        </CardBody>
      </Card>

      <Modal isOpen={isInviteModalOpen} onClose={onInviteModalClose}>
        <ModalContent>
          <ModalHeader>Пригласить сотрудника</ModalHeader>
          <ModalBody>
            <Select
              label="Роль"
              selectedKeys={[inviteRole]}
              onChange={(e) => setInviteRole(e.target.value as CoreUserRole)}
            >
              <SelectItem key={CoreUserRole.USER_ROLE_EMPLOYEE}>
                Сотрудник
              </SelectItem>
              <SelectItem key={CoreUserRole.USER_ROLE_ADMIN}>
                Администратор
              </SelectItem>
            </Select>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onInviteModalClose}>
              Отмена
            </Button>
            <Button color="success" isLoading={inviting} onPress={handleInvite}>
              Пригласить
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
      {isAdmin && (
        <Button
          isIconOnly
          className="fixed right-6 bottom-24 z-50 shadow-lg"
          color="secondary"
          isDisabled={inviting}
          isLoading={inviting}
          radius="full"
          size="lg"
          spinner={<Spinner color="success" size="sm" />}
          onPress={onInviteModalOpen}
        >
          <PlusIcon className="h-6 w-6" />
        </Button>
      )}
    </div>
  );
}
