export interface DocumentInfo {
  id: string;
  name: string;
  status: "processing" | "indexed" | "error";
  uploadDate: string; // ISO date string
  thumbnailUrl?: string; // Optional URL to a thumbnail image of the first page
}
