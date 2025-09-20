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
  wordState?: string;
  blockNumber?: number;
  blockHash?: string;
  merkleRoot?: string;
}

export interface ContractEventInBlock {
  contractId: string;
  eventId: string;
  type: string;
  actorId: string;
  payload: any;
  timestamp: string;
}

export interface Block {
  id: string;
  blockNumber: number;
  timestamp: string;
  contractEvents: ContractEventInBlock[];
  prevHash: string;
  hash: string;
  merkleRoot: string;
}

export interface LedgerResponse {
  transactions: Transaction[];
  blocks: Block[];
}
