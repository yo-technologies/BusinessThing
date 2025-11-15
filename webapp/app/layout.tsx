import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";
import { BottomNavbar } from "@/components/layout/BottomNavbar";
import { OrganizationSwitcher } from "@/components/layout/OrganizationSwitcher";
import { TelegramInit } from "@/components/TelegramInit";

export const metadata: Metadata = {
  title: {
    default: siteConfig.name,
    template: `%s - ${siteConfig.name}`,
  },
  description: siteConfig.description,
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "white" },
    { media: "(prefers-color-scheme: dark)", color: "black" },
  ],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html suppressHydrationWarning lang="ru">
      <head />
      <body
        className={clsx(
          "h-screen max-h-screen bg-background font-sans antialiased text-foreground pt-24",
          fontSans.variable,
        )}
      >
        <Providers themeProps={{ attribute: "class", defaultTheme: "dark" }}>
          <TelegramInit />
          <div className="flex h-full flex-col bg-gradient-to-b from-background via-background to-default-100">
            <OrganizationSwitcher />
            <main className="flex-1 px-3 pb-20 md:px-6">
              <div className="mx-auto flex h-full max-w-4xl flex-col">{children}</div>
            </main>
            <BottomNavbar />
          </div>
        </Providers>
      </body>
    </html>
  );
}
