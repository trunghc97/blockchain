export interface TransferRequest {
  fromUser: string;
  toAccount: string;
  amount: number;
  description: string;
  approvers: string[];
}

export interface TransferStatus {
  reqId: string;
  fromUser: string;
  toAccount: string;
  amount: number;
  description: string;
  status: 'PENDING' | 'PARTIALLY_APPROVED' | 'EXECUTED';
  approvers: string[];
  approvedBy: string[];
  createdAt: Date;
}

export interface ApproveRequest {
  reqId: string;
  approverId: string;
}