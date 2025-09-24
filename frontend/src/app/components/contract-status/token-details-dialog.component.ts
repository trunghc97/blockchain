import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { Contract } from '../../models/contract.model';

@Component({
  selector: 'app-token-details-dialog',
  template: `
    <div class="token-details-dialog">
      <h2 mat-dialog-title>Chi Tiết Token</h2>

      <mat-dialog-content>
        <div class="token-info mb-3">
          <div class="row">
            <div class="col-md-6">
              <strong>Token ID:</strong> {{ data.tokenId }}
            </div>
            <div class="col-md-6">
              <strong>Contract ID:</strong> {{ data.contract.contractId }}
            </div>
          </div>
          <div class="row mt-2">
            <div class="col-md-6">
              <strong>Total Supply:</strong> {{ data.contract.totalAmount | currency }}
            </div>
            <div class="col-md-6">
              <strong>Status:</strong>
              <span class="badge ms-2" [ngClass]="data.contract.approved ? 'bg-success' : 'bg-warning'">
                {{ data.contract.approved ? 'Approved' : 'Pending' }}
              </span>
            </div>
          </div>
        </div>

        <h4>Token Ownership</h4>
        <div class="table-responsive">
          <table class="table table-striped">
            <thead>
              <tr>
                <th>Account ID</th>
                <th>Account Type</th>
                <th>Balance</th>
                <th>Percentage</th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let balance of data.balances">
                <td>{{ balance.account }}</td>
                <td>{{ getAccountType(balance.account) }}</td>
                <td>{{ balance.balance | currency }}</td>
                <td>{{ getPercentage(balance.balance) }}%</td>
              </tr>
              <tr *ngIf="data.balances.length === 0">
                <td colspan="4" class="text-center text-muted">No balances found</td>
              </tr>
            </tbody>
          </table>
        </div>
      </mat-dialog-content>

      <mat-dialog-actions align="end">
        <button mat-button mat-dialog-close>Đóng</button>
      </mat-dialog-actions>
    </div>
  `,
  styles: [`
    .token-details-dialog {
      min-width: 500px;
    }

    .token-info {
      background-color: #f8f9fa;
      padding: 1rem;
      border-radius: 5px;
    }

    .badge {
      font-size: 0.8em;
    }
  `]
})
export class TokenDetailsDialogComponent {
  constructor(
    public dialogRef: MatDialogRef<TokenDetailsDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: {
      tokenId: string;
      contract: Contract;
      balances: any[];
    }
  ) {}

  getAccountType(accountId: string): string {
    if (accountId.startsWith('SUPPLIER')) {
      return 'Supplier';
    } else if (accountId.startsWith('ANCHOR')) {
      return 'Anchor';
    } else if (accountId.startsWith('BANK')) {
      return 'Bank';
    }
    return 'Unknown';
  }

  getPercentage(balance: number): string {
    if (this.data.contract.totalAmount && this.data.contract.totalAmount > 0) {
      return ((balance / this.data.contract.totalAmount) * 100).toFixed(1);
    }
    return '0.0';
  }
}
