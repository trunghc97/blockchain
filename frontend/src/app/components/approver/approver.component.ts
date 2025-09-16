import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { TransferService } from '../../services/transfer.service';
import { WorldState } from '../../models/transfer.model';

@Component({
  selector: 'app-approver',
  templateUrl: './approver.component.html',
  styleUrls: ['./approver.component.css']
})
export class ApproverComponent implements OnInit {
  approverForm: FormGroup;
  worldState?: WorldState;
  error?: string;

  constructor(
    private fb: FormBuilder,
    private transferService: TransferService
  ) {
    this.approverForm = this.fb.group({
      transactionId: ['', Validators.required],
      approverId: ['', Validators.required]
    });
  }

  ngOnInit() {}

  onCheckStatus() {
    const transactionId = this.approverForm.get('transactionId')?.value;
    if (transactionId) {
      this.transferService.getTransferStatus(transactionId).subscribe({
        next: (response) => {
          this.worldState = response;
          this.error = undefined;
        },
        error: (err) => {
          this.error = 'Error fetching transaction status';
          this.worldState = undefined;
        }
      });
    }
  }

  onApprove() {
    if (this.approverForm.valid && this.worldState) {
      const request = {
        transactionId: this.worldState.transactionId,
        fromAccount: this.worldState.fromAccount,
        toAccount: this.worldState.toAccount,
        amount: this.worldState.amount,
        approverId: this.approverForm.get('approverId')?.value
      };

      this.transferService.approveTransfer(request).subscribe({
        next: (response) => {
          this.worldState = response;
          this.error = undefined;
        },
        error: (err) => {
          this.error = 'Error approving transaction';
        }
      });
    }
  }
}
