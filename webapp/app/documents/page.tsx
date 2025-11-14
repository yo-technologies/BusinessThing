"use client";

import { useState } from "react";
import { DocumentCard } from "@/components/documents/DocumentCard";
import { DocumentInfo } from "@/types/document";
import { Button } from "@heroui/button";
import { PlusIcon } from "@heroicons/react/24/outline";

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

  const handleDelete = (id: string) => {
    console.log(`Deleting document with ID: ${id}`);
    setDocuments((prevDocs) => prevDocs.filter((doc) => doc.id !== id));
  };

  // For now, we'll assume isAdmin is true for demonstration purposes
  const isAdmin = true;

  return (
    <div className="p-4">
      <h1 className="text-3xl font-bold mb-4">Documents</h1>

      <div className="flex justify-end mb-4 w-full">
        <Button size="sm" startContent={<PlusIcon className="w-5 h-5" />} className="w-full">
          Upload Document
        </Button>
      </div>

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
    </div>
  );
}
