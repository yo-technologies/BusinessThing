"use client";

import "@/styles/globals.css";
import clsx from "clsx";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";
import { BottomNavbar } from "@/components/layout/BottomNavbar";
import { OrganizationSwitcher } from "@/components/layout/OrganizationSwitcher";
import { useTelegramViewport } from "@/hooks/useTelegramViewport";

export function ClientLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isFullscreen } = useTelegramViewport();

  return (
    <html suppressHydrationWarning lang="ru">
      <head />
      <body
        className={clsx(
          "h-screen bg-background font-sans antialiased text-foreground overflow-hidden",
          fontSans.variable,
        )}
      >
        <Providers themeProps={{ attribute: "class", defaultTheme: "dark" }}>
          <div className="flex h-screen flex-col ">
            <header className="shrink-0 px-3 py-2">
              <OrganizationSwitcher />
            </header>
            <main className="flex-1 overflow-y-auto">
              <div 
                className={clsx(
                  "mx-auto flex h-full max-w-4xl flex-col px-4 mb-23",
                  isFullscreen && "mt-25"
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
