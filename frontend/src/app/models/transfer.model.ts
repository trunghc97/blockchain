export interface TransferRequest {
  transactionId?: string;  // Optional vì sẽ được tạo tự động
  fromAccount: string;     // Đổi tên từ fromUser
  toAccount: string;
  amount: number;
  approvers: string[];
}

export interface Approver {
  userId: string;
  status: string;
  timestamp: Date;
}

export interface TransferStatus {
  id: string;
  transactionId: string;
  fromAccount: string;
  toAccount: string;
  amount: number;
  status: 'PENDING' | 'PARTIALLY_APPROVED' | 'APPROVED' | 'APPROVED_PENDING_EXEC' | 'EXECUTED';
  approvers: Approver[];
  approvalCount: number;
  supplierRef?: string;
  lastUpdated: Date;
}

export interface ApproveRequest {
  transactionId: string;
  approverUserId: string;
  fromAccount: string;
  toAccount: string;
  amount: number;
}
