import "@/styles/globals.css";
import { Metadata, Viewport } from "next";
import clsx from "clsx";

import { Providers } from "./providers";

import { siteConfig } from "@/config/site";
import { fontSans } from "@/config/fonts";
import { BottomNavbar } from "@/components/layout/BottomNavbar";
import { OrganizationSwitcher } from "@/components/layout/OrganizationSwitcher";

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
          <div className="flex min-h-screen flex-col bg-gradient-to-b from-background via-background to-default-100">
            <div className="fixed left-0 right-0 top-0 z-40 flex items-center justify-between border-b border-default-200 bg-background/80 px-4 py-2 backdrop-blur-md">
              <OrganizationSwitcher />
            </div>
            <main className="flex-1 px-3 pb-20 pt-16 md:px-6 md:pb-24">
              <div className="mx-auto flex h-full max-w-4xl flex-col">{children}</div>
            </main>
            <BottomNavbar />
          </div>
        </Providers>
      </body>
    </html>
  );
}
