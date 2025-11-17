"use client";

import "@/styles/globals.css";
import clsx from "clsx";

import { Providers } from "./providers";

import { fontSans } from "@/config/fonts";
import { BottomNavbar } from "@/components/layout/BottomNavbar";
import { OrganizationSwitcher } from "@/components/layout/OrganizationSwitcher";
import { useTelegramViewport } from "@/hooks/useTelegramViewport";

export function ClientLayout({ children }: { children: React.ReactNode }) {
  const { isFullscreen } = useTelegramViewport();

  return (
    <html suppressHydrationWarning lang="ru">
      <head />
      <body
        className={clsx(
          "bg-background font-sans antialiased text-foreground",
          fontSans.variable,
        )}
      >
        <Providers themeProps={{ attribute: "class", defaultTheme: "dark" }}>
          <div className="fixed t-0 w-screen z-100 h-15 flex place-content-center place-items-center">
            <OrganizationSwitcher />
          </div>
          <div className="flex h-screen flex-col">
            <main className="flex-1 overflow-auto">
              <div
                className={clsx(
                  "mx-auto flex h-full max-w-4xl flex-col px-4 pb-22",
                  isFullscreen ? "pt-22" : "pt-15",
                )}
              >
                {children}
              </div>
            </main>
            <BottomNavbar />
          </div>
        </Providers>
      </body>
    </html>
  );
}
