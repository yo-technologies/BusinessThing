"use client";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col min-h-full px-4">
      {children}
    </div>
  );
}
