"use client";

import { useState } from "react";
import { DocumentCard } from "@/components/documents/DocumentCard";
import { DocumentInfo } from "@/types/document";
import { Button } from "@heroui/button";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useAuth } from "@/hooks/useAuth"; // Import useAuth hook

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<DocumentInfo[]>([
    {
      id: "1",
      name: "Project Proposal.pdf",
      status: "indexed",
      uploadDate: "2023-10-26T10:00:00Z",
      thumbnailUrl: "/makan.png",
    },
    {
      id: "2",
      name: "Meeting Notes.docx",
      status: "processing",
      uploadDate: "2023-10-25T14:30:00Z",
      thumbnailUrl: undefined, // No thumbnail for this one
    },
    {
      id: "3",
      name: "Legal Contract.txt",
      status: "error",
      uploadDate: "2023-10-24T09:15:00Z",
      thumbnailUrl: "/makan.png",
    },
    {
      id: "4",
      name: "Financial Report.pdf",
      status: "indexed",
      uploadDate: "2023-10-23T11:45:00Z",
      thumbnailUrl: "/makan.png",
    },
  ]);

  const { isAdmin, loading } = useAuth(); // Use the useAuth hook

  const handleDelete = (id: string) => {
    console.log(`Deleting document with ID: ${id}`);
    setDocuments((prevDocs) => prevDocs.filter((doc) => doc.id !== id));
  };

  if (loading) {
    return <div>Loading authentication...</div>; // Or a spinner
  }
  console.log(isAdmin)

  return (
    <div className="p-4">
      <h1 className="text-3xl font-bold mb-4">База знаний</h1> {/* Renamed title */}

      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {documents.map((doc) => (
          <DocumentCard
            key={doc.id}
            documentInfo={doc}
            onDelete={handleDelete}
            isAdmin={isAdmin}
          />
        ))}
      </div>

      {isAdmin && ( // Conditionally render upload button for admin
        <div className="fixed bottom-20 right-6 z-40"> {/* Fixed position, bottom-right */}
          <Button isIconOnly color="primary" size="lg" radius="full" className="w-16 h-16 md:w-20 sm:h-20 ring-2 ring-white">
            <PlusIcon className="w-7 h-7" />
          </Button>
        </div>
      )}
    </div>
  );
}
