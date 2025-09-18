import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TransferService } from '../../services/transfer.service';
import { TransferStatus, ApproveRequest } from '../../models/transfer.model';

@Component({
  selector: 'app-approver',
  templateUrl: './approver.component.html',
  styleUrls: ['./approver.component.css']
})
export class ApproverComponent implements OnInit {
  pendingTransfers: TransferStatus[] = [];
  displayedColumns = ['transactionId', 'fromAccount', 'toAccount', 'amount', 'status', 'approvalCount', 'actions'];
  isApproved = (transfer: TransferStatus) => {
    return transfer.approvers?.some(a => a.userId === this.currentUserId && a.status === 'APPROVED');
  };
  currentUserId: string = '';

  constructor(
    private transferService: TransferService,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit(): void {
    this.currentUserId = localStorage.getItem('currentUserId') || '';
    this.loadPendingTransfers();
  }

  loadPendingTransfers(): void {
    if (this.currentUserId) {
      this.transferService.getPendingTransfers(this.currentUserId).subscribe(
        transfers => {
          this.pendingTransfers = transfers;
        },
        error => {
          this.snackBar.open('Có lỗi khi tải danh sách giao dịch', 'Đóng', {
            duration: 3000
          });
        }
      );
    }
  }

  onApprove(transfer: TransferStatus): void {
    if (this.currentUserId) {
      const approveRequest: ApproveRequest = {
        transactionId: transfer.transactionId,
        approverUserId: this.currentUserId,
        fromAccount: transfer.fromAccount,
        toAccount: transfer.toAccount,
        amount: transfer.amount
      };
      this.transferService.approveTransfer(approveRequest).subscribe(
        () => {
          this.snackBar.open('Phê duyệt thành công', 'Đóng', {
            duration: 3000
          });
          this.loadPendingTransfers();
        },
        error => {
          this.snackBar.open('Có lỗi khi phê duyệt giao dịch', 'Đóng', {
            duration: 3000
          });
        }
      );
    }
  }
}
