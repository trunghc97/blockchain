export interface TransferRequest {
    transactionId: string;
    fromAccount: string;
    toAccount: string;
    amount: number;
    approverId?: string;
}

export interface WorldState {
    id: string;
    transactionId: string;
    fromAccount: string;
    toAccount: string;
    amount: number;
    status: string;
    approvalCount: number;
    lastUpdated: string;
}
