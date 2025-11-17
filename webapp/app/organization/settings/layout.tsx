"use client";

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col h-full px-4">
      {children}
      <div className="pb-23" />
    </div>
  );
}
