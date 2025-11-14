import { Card, CardHeader, CardBody, CardFooter } from "@heroui/card";
import { Chip } from "@heroui/chip";
import { Image } from "@heroui/image";
import { Button } from "@heroui/button";
import { DocumentInfo } from "@/types/document";
import { TrashIcon, DocumentIcon } from "@heroicons/react/24/outline";

interface DocumentCardProps {
  documentInfo: DocumentInfo;
  onDelete?: (id: string) => void; // Optional delete handler for Admin
  isAdmin?: boolean; // To show delete button for Admin
}

export const DocumentCard = ({ documentInfo, onDelete, isAdmin }: DocumentCardProps) => {
  const getStatusColor = (status: DocumentInfo["status"]) => {
    switch (status) {
      case "indexed":
        return "success";
      case "processing":
      case "error": // Treat error as warning for color
        return "warning";
      default:
        return "default";
    }
  };

  return (
    <Card className="py-0 group relative overflow-hidden h-50">
      {documentInfo.thumbnailUrl ? (
        <Image
          alt="Document thumbnail"
          className="absolute inset-0 object-cover"
          src={documentInfo.thumbnailUrl}
          width={1000}
        />
      ) : (
        <div className="absolute inset-0 flex items-center justify-center">
          <DocumentIcon className="w-16 h-16 text-default-400" />
        </div>
      )}
      {/* Overlay for better text readability */}
      

      <CardHeader className="absolute top-0 left-0 right-0 pb-0 pt-2 px-4 flex-col items-start z-20">
        <div className="flex gap-1 mt-2">
          <Chip size="sm" className="text-white bg-zinc-700">
            {new Date(documentInfo.uploadDate).toLocaleDateString()}
          </Chip>
          <Chip color={getStatusColor(documentInfo.status)} size="sm" className="text-white">
            {documentInfo.status}
          </Chip>
        </div>
      </CardHeader>
      <CardBody className="absolute inset-0 flex items-center justify-center z-20"></CardBody>
      
      {isAdmin && onDelete && (
        <div className="absolute top-2 right-2 z-30">
          <Button
            isIconOnly
            color="danger"
            variant="solid"
            size="sm"
            onPress={() => onDelete(documentInfo.id)}
          >
            <TrashIcon className="w-4 h-4" />
          </Button>
        </div>
      )}

      <CardFooter className="absolute bottom-0 left-0 right-0 px-4 pb-2 z-20 justify-between backdrop-blur-md bg-default/70">
        <h4 className="font-bold text-large">{documentInfo.name}</h4>
      </CardFooter>
    </Card>
  );
};
