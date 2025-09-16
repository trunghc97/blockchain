import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TransferService } from '../../services/transfer.service';
import { TransferStatus } from '../../models/transfer.model';

@Component({
  selector: 'app-approver',
  templateUrl: './approver.component.html',
  styleUrls: ['./approver.component.css']
})
export class ApproverComponent implements OnInit {
  pendingTransfers: TransferStatus[] = [];
  displayedColumns: string[] = ['reqId', 'fromUser', 'toAccount', 'amount', 'description', 'status', 'actions'];
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

  onApprove(reqId: string): void {
    if (this.currentUserId) {
      this.transferService.approveTransfer({ reqId, approverId: this.currentUserId }).subscribe(
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