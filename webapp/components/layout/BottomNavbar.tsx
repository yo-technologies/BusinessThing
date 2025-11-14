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
import { clsx } from "clsx";

const navigationItems = [
  {
    href: "/chat",
    label: "Chat",
    Icon: ChatBubbleOvalLeftEllipsisIcon,
    SolidIcon: SolidChatBubbleIcon,
  },
  {
    href: "/documents",
    label: "Documents",
    Icon: DocumentTextIcon,
    SolidIcon: SolidDocumentTextIcon,
  },
  {
    href: "/generator",
    label: "Generator",
    Icon: DocumentDuplicateIcon,
    SolidIcon: SolidDocumentDuplicateIcon,
  },
  {
    href: "/management",
    label: "Management",
    Icon: Cog6ToothIcon,
    SolidIcon: SolidCog6ToothIcon,
    adminOnly: true, // We can use this later for role-based rendering
  },
];

export const BottomNavbar = () => {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-background border-t border-default-200">
      <div className="flex justify-around max-w-lg mx-auto">
        {navigationItems.map(({ href, label, Icon, SolidIcon }) => {
          const isActive = pathname.startsWith(href);
          const CurrentIcon = isActive ? SolidIcon : Icon;

          return (
            <Link
              key={href}
              href={href}
              className={clsx(
                "flex flex-col items-center justify-center w-full pt-2 pb-1 text-sm",
                {
                  "text-primary": isActive,
                  "text-foreground-500 hover:text-foreground-700": !isActive,
                }
              )}
            >
              <CurrentIcon className="w-6 h-6" />
              <span>{label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
};
