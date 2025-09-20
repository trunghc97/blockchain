export interface SupplierAmount {
  supplierId: string;
  name: string;
  amount: number; // Changed from allocatedAmount to match backend
  status: string;
}

export interface Contract {
  id: string;
  contractId: string;
  description: string;
  buyer: string; // Changed from buyer: string to match backend field
  suppliers: SupplierAmount[];
  totalAmount: number;
  status: string;
  fileUrl?: string;
  createdAt: Date;
  updatedAt: Date;
  wordState?: string;
}