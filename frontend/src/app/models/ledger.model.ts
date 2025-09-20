export interface Transaction {
  id: string;
  contractId: string;
  type: string;
  buyer: string;
  bank: string;
  suppliers: any[];
  totalAmount: number;
  description: string;
  approverID?: string;
  status: string;
  timestamp: string;
  included: boolean;
}

export interface Block {
  id: string;
  blockNumber: number;
  timestamp: string;
  previousHash: string;
  hash: string;
  txIds: string[];
}

export interface LedgerResponse {
  transactions: Transaction[];
  blocks: Block[];
}
