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
    <nav className="fixed bottom-0 left-0 right-0 z-50 pt-3">
      <div className="mx-auto max-w-lg px-4 pb-4">
        <div className="bg-black/5 backdrop-blur-xs rounded-full px-1 py-1 grid grid-cols-4 gap-0 border border-white/15">
          {navigationItems.map(({ href, label, Icon, SolidIcon }) => {
            const isActive = pathname.startsWith(href);
            const CurrentIcon = isActive ? SolidIcon : Icon;

            return (
              <Link
                key={href}
                href={href}
                className={clsx(
                  "flex flex-col items-center justify-center gap-1 px-3 py-1 rounded-full text-xs font-medium transition-all duration-300",
                  {
                    "bg-zinc-800/70 text-primary-500 shadow-lg": isActive,
                    "text-zinc-400 hover:text-zinc-200 hover:bg-white/5": !isActive,
                  },
                )}
              >
                <CurrentIcon className="h-7 w-7" />
                <span className="text-[9px] whitespace-nowrap font-semibold">{label}</span>
              </Link>
            );
          })}
        </div>
      </div>
    </nav>
  );
};
