"use client";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col h-full">
      {children}
    </div>
  );
}
