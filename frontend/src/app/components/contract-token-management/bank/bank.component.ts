import { Component, OnInit } from '@angular/core';
import { ContractTokenService } from '../../../services/contract-token.service';
import { ContractService } from '../../../services/contract.service';

@Component({
  selector: 'app-bank',
  templateUrl: './bank.component.html',
  styleUrls: ['./bank.component.css']
})
export class BankComponent implements OnInit {
  contracts: any[] = [];
  tokens: any[] = [];
  balances: any[] = [];
  loading = {
    contracts: false,
    tokens: false,
    balances: false
  };
  activeTab = 'contracts';
  bankId = 'BANK001'; // This should come from authentication context

  constructor(
    private contractTokenService: ContractTokenService,
    private contractService: ContractService
  ) {}

  ngOnInit(): void {
    this.loadAllData();
  }

  loadAllData(): void {
    this.loadContracts();
    this.loadAllTokens();
    this.loadAllBalances();
  }

  loadContracts(): void {
    this.loading.contracts = true;
    this.contractService.getContracts().subscribe({
      next: (contracts) => {
        this.contracts = contracts;
        this.loading.contracts = false;
      },
      error: (error) => {
        console.error('Error loading contracts:', error);
        alert('Error loading contracts: ' + error.message);
        this.loading.contracts = false;
      }
    });
  }

  loadAllTokens(): void {
    this.loading.tokens = true;
    this.contractService.getAllTokens().subscribe({
      next: (tokens) => {
        this.tokens = tokens;
        this.loading.tokens = false;
      },
      error: (error) => {
        console.error('Error loading tokens:', error);
        alert('Error loading tokens: ' + error.message);
        this.loading.tokens = false;
      }
    });
  }

  loadAllBalances(): void {
    this.loading.balances = true;
    this.contractService.getAllBalances().subscribe({
      next: (balances) => {
        this.balances = balances;
        this.loading.balances = false;
      },
      error: (error) => {
        console.error('Error loading balances:', error);
        alert('Error loading balances: ' + error.message);
        this.loading.balances = false;
      }
    });
  }

  approveContract(contract: any): void {
    if (confirm(`Are you sure you want to approve contract ${contract.contractId}?`)) {
      this.contractService.approveContractByBank(contract.contractId, { bankId: this.bankId }).subscribe({
        next: (result) => {
          alert('Contract approved successfully!');
          this.loadContracts();
        },
        error: (error) => {
          console.error('Error approving contract:', error);
          alert('Error approving contract: ' + error.message);
        }
      });
    }
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

  getContractStatus(contract: any): string {
    if (contract.status === 'PENDING_BANK_APPROVAL') {
      return 'Pending Bank Approval';
    } else if (contract.status === 'BANK_APPROVED') {
      return 'Bank Approved';
    } else if (contract.status === 'EXECUTED') {
      return 'Executed';
    } else {
      return contract.status;
    }
  }

  getContractStatusBadgeClass(contract: any): string {
    switch (contract.status) {
      case 'PENDING_BANK_APPROVAL': return 'bg-warning';
      case 'BANK_APPROVED': return 'bg-info';
      case 'EXECUTED': return 'bg-success';
      default: return 'bg-secondary';
    }
  }

  canApproveContract(contract: any): boolean {
    return contract.status === 'PENDING_BANK_APPROVAL' && contract.bankId === this.bankId;
  }

  get transferredTokensCount(): number {
    return this.tokens.filter(t => t.owner !== this.bankId).length;
  }

  get heldTokensCount(): number {
    return this.tokens.filter(t => t.owner === this.bankId).length;
  }

  get pendingContractsCount(): number {
    return this.contracts.filter(c => c.status === 'PENDING_BANK_APPROVAL').length;
  }

  get approvedContractsCount(): number {
    return this.contracts.filter(c => c.status === 'BANK_APPROVED' || c.status === 'EXECUTED').length;
  }

  get totalTokensValue(): number {
    return this.tokens.reduce((sum, token) => sum + (token.total || 0), 0);
  }

  get totalBalancesValue(): number {
    return this.balances.reduce((sum, balance) => sum + (balance.balance || 0), 0);
  }
}
