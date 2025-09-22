import { Component, OnInit } from '@angular/core';
import { ContractTokenService } from '../../../services/contract-token.service';

@Component({
  selector: 'app-bank',
  templateUrl: './bank.component.html',
  styleUrls: ['./bank.component.css']
})
export class BankComponent implements OnInit {
  tokens: any[] = [];
  loading = false;
  bankId = 'BANK001'; // This should come from authentication context

  constructor(private contractTokenService: ContractTokenService) {}

  ngOnInit(): void {
    this.loadTokens();
  }

  loadTokens(): void {
    this.loading = true;
    this.contractTokenService.getTokensIssuedByBank(this.bankId).subscribe({
      next: (tokens) => {
        this.tokens = tokens;
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading tokens:', error);
        alert('Error loading tokens: ' + error.message);
        this.loading = false;
      }
    });
  }

  getTokenStatus(token: any): string {
    if (token.owner === this.bankId) {
      return 'Held by Bank';
    } else {
      return `Transferred to ${token.owner}`;
    }
  }

  getStatusBadgeClass(token: any): string {
    return token.owner === this.bankId ? 'bg-primary' : 'bg-success';
  }

  get transferredTokensCount(): number {
    return this.tokens.filter(t => t.owner !== this.bankId).length;
  }

  get heldTokensCount(): number {
    return this.tokens.filter(t => t.owner === this.bankId).length;
  }
}
