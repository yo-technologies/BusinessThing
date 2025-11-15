"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ChatBubbleOvalLeftEllipsisIcon,
  DocumentTextIcon,
  DocumentDuplicateIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";
import {
  ChatBubbleOvalLeftEllipsisIcon as SolidChatBubbleIcon,
  DocumentTextIcon as SolidDocumentTextIcon,
  DocumentDuplicateIcon as SolidDocumentDuplicateIcon,
  Cog6ToothIcon as SolidCog6ToothIcon,
} from "@heroicons/react/24/solid";
import clsx from "clsx";

const navigationItems = [
  {
    href: "/chat",
    label: "Чат",
    Icon: ChatBubbleOvalLeftEllipsisIcon,
    SolidIcon: SolidChatBubbleIcon,
  },
  {
    href: "/documents",
    label: "База знаний",
    Icon: DocumentTextIcon,
    SolidIcon: SolidDocumentTextIcon,
  },
  {
    href: "/generator",
    label: "Генератор",
    Icon: DocumentDuplicateIcon,
    SolidIcon: SolidDocumentDuplicateIcon,
  },
  {
    href: "/management",
    label: "Управление",
    Icon: Cog6ToothIcon,
    SolidIcon: SolidCog6ToothIcon,
    adminOnly: true, // We can use this later for role-based rendering
  },
];

export const BottomNavbar = () => {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-default-200 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-lg justify-around px-2 py-1 mb-2">
        {navigationItems.map(({ href, label, Icon, SolidIcon }) => {
          const isActive = pathname.startsWith(href);
          const CurrentIcon = isActive ? SolidIcon : Icon;

          return (
            <Link
              key={href}
              href={href}
              className={clsx(
                "flex w-full flex-col items-center justify-center gap-0.5 rounded-full px-1 py-1 text-[11px] font-medium transition-colors",
                {
                  "text-primary bg-primary-50": isActive,
                  "text-default-500 hover:text-default-700": !isActive,
                },
              )}
            >
              <CurrentIcon className="h-5 w-5" />
              <span className="text-[11px]">{label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
};
