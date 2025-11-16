"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card, CardBody, CardHeader } from "@heroui/card";
import { Spinner } from "@heroui/spinner";
import { Chip } from "@heroui/chip";
import { Modal, ModalBody, ModalContent, ModalFooter, ModalHeader, useDisclosure } from "@heroui/modal";
import { Input } from "@heroui/input";
import { Select, SelectItem } from "@heroui/select";
import { Tabs, Tab } from "@heroui/tabs";
import {
  UserPlusIcon,
  UserIcon,
  ClipboardDocumentCheckIcon,
  ShieldCheckIcon,
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  EnvelopeIcon,
  UsersIcon,
} from "@heroicons/react/24/outline";
import { useRouter } from "next/navigation";

import { CoreUser, CoreUserRole, CoreUserStatus, CoreInvitation } from "@/api/api.core.generated";
import { useApiClients } from "@/api/client";
import { useAuth } from "@/hooks/useAuth";
import { useOrganization } from "@/hooks/useOrganization";
import { useBackButton } from "@/hooks/useBackButton";
import { useCurrentRole } from "@/hooks/useCurrentRole";

type TabType = "users" | "invitations";

export default function UsersPage() {
  const router = useRouter();
  const { loading: authLoading, isAuthenticated, isNewUser, organizations, userInfo } = useAuth();
  const { currentOrg, loading: orgLoading, needsOrganization } = useOrganization({ organizations });
  const { core } = useApiClients();
  const { isAdmin } = useCurrentRole();

  const [activeTab, setActiveTab] = useState<TabType>("users");
  const [users, setUsers] = useState<CoreUser[]>([]);
  const [invitations, setInvitations] = useState<CoreInvitation[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [inviteRole, setInviteRole] = useState<CoreUserRole>(CoreUserRole.USER_ROLE_EMPLOYEE);
  const [inviting, setInviting] = useState(false);
  const [invitationUrl, setInvitationUrl] = useState<string | null>(null);

  const { isOpen: isInviteModalOpen, onOpen: onInviteModalOpen, onClose: onInviteModalClose } = useDisclosure();
  const { isOpen: isLinkModalOpen, onOpen: onLinkModalOpen, onClose: onLinkModalClose } = useDisclosure();

  useBackButton(true);

  useEffect(() => {
    if (!authLoading && isNewUser) {
      router.replace("/onboarding");
    }
  }, [isNewUser, authLoading, router]);

  useEffect(() => {
    if (!authLoading && !orgLoading && isAuthenticated && !isNewUser && needsOrganization) {
      router.replace("/organization/create");
    }
  }, [authLoading, orgLoading, isAuthenticated, isNewUser, needsOrganization, router]);

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
  }, [isAuthenticated, authLoading, isNewUser, currentOrg?.id, loadUsers, loadInvitations]);

  const handleInvite = async () => {
    if (!currentOrg?.id) return;

    setInviting(true);
    try {
      const response = await core.v1.userServiceInviteUser(currentOrg.id, {
        role: inviteRole,
      });
      if (response.data.invitationUrl) {
        setInvitationUrl(response.data.invitationUrl);
        onInviteModalClose();
        onLinkModalOpen();
        void loadInvitations();
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

  const getRoleLabel = (role?: CoreUserRole) => {
    switch (role) {
      case CoreUserRole.USER_ROLE_ADMIN:
        return { label: "Администратор", color: "secondary" as const, icon: ShieldCheckIcon };
      case CoreUserRole.USER_ROLE_EMPLOYEE:
        return { label: "Сотрудник", color: "default" as const, icon: UserIcon };
      default:
        return { label: "Неизвестно", color: "default" as const, icon: UserIcon };
    }
  };

  const getStatusLabel = (status?: CoreUserStatus) => {
    switch (status) {
      case CoreUserStatus.USER_STATUS_ACTIVE:
        return { label: "Активен", color: "success" as const, icon: CheckCircleIcon };
      case CoreUserStatus.USER_STATUS_PENDING:
        return { label: "Ожидает", color: "warning" as const, icon: ClockIcon };
      case CoreUserStatus.USER_STATUS_INACTIVE:
        return { label: "Неактивен", color: "danger" as const, icon: XCircleIcon };
      default:
        return { label: "Неизвестно", color: "default" as const, icon: ClockIcon };
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
        aria-label="Управление командой"
        selectedKey={activeTab}
        onSelectionChange={(key) => setActiveTab(key as TabType)}
        classNames={{
          tabList: "w-full relative rounded-full p-0",
          cursor: "rounded-full",
          panel: "p-0",
        }}
        fullWidth
        size="lg"
      >
        <Tab
          key="users"
          title={
            <div className="flex items-center gap-2 rounded-full">
              <UsersIcon className="h-4 w-4" />
              <span>Пользователи</span>
              <Chip size="sm" variant="flat" color="secondary">
                {users.length}
              </Chip>
            </div>
          }
        >
          {users.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-2">
              <UserIcon className="h-16 w-16 text-default-300" />
              <p className="text-default-400 text-center">Нет сотрудников</p>
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
                            <Chip size="sm" variant="flat" color="default">
                              Вы
                            </Chip>
                          )}
                        </div>
                        <p className="text-sm text-default-400 truncate">{user.email || "Нет email"}</p>
                      </div>
                    </div>
                    <div className="flex gap-2 flex-wrap">
                      <Chip
                        color={role.color}
                        variant="flat"
                        size="sm"
                        startContent={<RoleIcon className="h-4 w-4" />}
                      >
                        {role.label}
                      </Chip>
                      <Chip
                        color={status.color}
                        variant="flat"
                        size="sm"
                        startContent={<StatusIcon className="h-4 w-4" />}
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
          title={
            <div className="flex items-center gap-2">
              <EnvelopeIcon className="h-4 w-4" />
              <span>Приглашения</span>
              <Chip size="sm" variant="flat" color="warning">
                {invitations.length}
              </Chip>
            </div>
          }
        >
          {invitations.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-2">
              <EnvelopeIcon className="h-16 w-16 text-default-300" />
              <p className="text-default-400 text-center">Нет активных приглашений</p>
              <p className="text-xs text-default-300 text-center max-w-xs">
                Создайте приглашение, чтобы добавить новых участников в команду
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {invitations.map((invitation) => {
                const role = getRoleLabel(invitation.role);
                const RoleIcon = role.icon;
                const isExpired = invitation.expiresAt ? new Date(invitation.expiresAt) < new Date() : false;
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
                        <p className="font-semibold">Приглашение #{invitation.id?.slice(0, 8)}</p>
                        <p className="text-xs text-default-400">
                          Создано: {invitation.createdAt ? new Date(invitation.createdAt).toLocaleDateString('ru-RU') : 'N/A'}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2 flex-wrap">
                      <Chip
                        color={role.color}
                        variant="flat"
                        size="sm"
                        startContent={<RoleIcon className="h-4 w-4" />}
                      >
                        {role.label}
                      </Chip>
                      {isUsed ? (
                        <Chip color="success" variant="flat" size="sm" startContent={<CheckCircleIcon className="h-4 w-4" />}>
                          Использовано
                        </Chip>
                      ) : isExpired ? (
                        <Chip color="danger" variant="flat" size="sm" startContent={<XCircleIcon className="h-4 w-4" />}>
                          Истекло
                        </Chip>
                      ) : (
                        <Chip color="warning" variant="flat" size="sm" startContent={<ClockIcon className="h-4 w-4" />}>
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
            <Button
              color="success"
              isLoading={inviting}
              onPress={handleInvite}
            >
              Пригласить
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal isOpen={isLinkModalOpen} onClose={onLinkModalClose}>
        <ModalContent>
          <ModalHeader>Приглашение создано</ModalHeader>
          <ModalBody>
            <p className="text-sm text-default-600 mb-2">
              Отправьте эту ссылку приглашения пользователю:
            </p>
            <div className="p-3 bg-default-100 rounded-lg break-all text-sm">
              {invitationUrl}
            </div>
            <p className="text-xs text-default-400 mt-2">
              Ссылка действительна в течение ограниченного времени
            </p>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onLinkModalClose}>
              Закрыть
            </Button>
            <Button
              color="secondary"
              startContent={<ClipboardDocumentCheckIcon className="h-5 w-5" />}
              onPress={handleCopyInvitation}
            >
              Скопировать
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}

