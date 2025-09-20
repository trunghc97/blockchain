export interface Contract {
    contractId?: string;
    description: string;
    suppliers: SupplierAmount[];
    file?: File;
    status?: string;
    createdAt?: Date;
    updatedAt?: Date;
}

export interface SupplierAmount {
    supplierId: string;
    amount: number;
}