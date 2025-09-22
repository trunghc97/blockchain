import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ContractTokenService } from '../../../services/contract-token.service';

@Component({
  selector: 'app-supplier',
  templateUrl: './supplier.component.html',
  styleUrls: ['./supplier.component.css']
})
export class SupplierComponent implements OnInit {
  tokens: any[] = [];
  transferForm: FormGroup;
  loading = false;
  supplierId = 'SUPPLIER001'; // This should come from authentication context

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService
  ) {
    this.transferForm = this.fb.group({
      tokenId: ['', Validators.required],
      to: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]]
    });
  }

  ngOnInit(): void {
    this.loadMyTokens();
  }

  loadMyTokens(): void {
    this.loading = true;
    // This is a simplified approach - in real implementation,
    // we'd need an API to get tokens by owner
    // For now, we'll simulate by getting all tokens and filtering
    this.contractTokenService.getAllTokens().subscribe({
      next: (allTokens) => {
        this.tokens = allTokens.filter((token: any) => token.owner === this.supplierId);
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading tokens:', error);
        alert('Error loading tokens: ' + error.message);
        this.loading = false;
      }
    });
  }

  onTransfer(): void {
    if (this.transferForm.valid) {
      this.loading = true;
      const transferData = {
        tokenId: this.transferForm.value.tokenId,
        from: this.supplierId,
        to: this.transferForm.value.to,
        amount: this.transferForm.value.amount
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token transferred:', response);
          alert('Token transferred successfully!');
          this.transferForm.reset();
          this.loadMyTokens();
        },
        error: (error) => {
          console.error('Error transferring token:', error);
          alert('Error transferring token: ' + error.message);
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  onPayBack(token: any): void {
    if (confirm(`Are you sure you want to return this token to the bank?`)) {
      this.loading = true;
      const transferData = {
        tokenId: token.id,
        from: this.supplierId,
        to: token.issuer, // Bank ID
        amount: token.total
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token returned to bank:', response);
          alert('Token returned to bank successfully!');
          this.loadMyTokens();
        },
        error: (error) => {
          console.error('Error returning token to bank:', error);
          alert('Error returning token to bank: ' + error.message);
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  get availableTokens(): any[] {
    return this.tokens.filter(token => token.owner === this.supplierId);
  }
}
