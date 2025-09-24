import { Component, OnInit } from '@angular/core';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatSelectChange } from '@angular/material/select';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';
import { firstValueFrom } from 'rxjs';
import { Contract } from '../../models/contract.model';
import { ContractService } from '../../services/contract.service';
import { ContractTokenService } from '../../services/contract-token.service';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user.model';

// Import the dialog component we'll create
import { TokenDetailsDialogComponent } from './token-details-dialog.component';

@Component({
  selector: 'app-contract-status',
  templateUrl: './contract-status.component.html',
  styleUrls: ['./contract-status.component.css']
})
export class ContractStatusComponent implements OnInit {
  contracts: Contract[] = [];
  filteredContracts: Contract[] = [];
  loading = false;
  expandedContracts: { [key: string]: boolean } = {};
  currentUser: User | null = null;

  // Token information for suppliers
  contractTokens: { [contractId: string]: any } = {};
  supplierTokens: { [supplierId: string]: any[] } = {};

  constructor(
    private contractService: ContractService,
    private contractTokenService: ContractTokenService,
    private userService: UserService,
    private snackBar: MatSnackBar,
    private dialog: MatDialog
  ) {}

  ngOnInit() {
    this.loadCurrentUserAndContracts();
  }

  async loadCurrentUserAndContracts() {
    try {
      this.currentUser = await firstValueFrom(this.userService.getCurrentUser());
      console.log('Current user:', this.currentUser);

      await this.loadContracts();
    } catch (error) {
      console.error('Error loading current user:', error);
    }
  }

  async loadContracts() {
    this.loading = true;
    try {
      this.contracts = await firstValueFrom(this.contractService.getContracts());
      this.filteredContracts = [...this.contracts];
      console.log('Loaded contracts with supplier names:', this.contracts);
    } catch (error) {
      console.error('Error loading contracts:', error);
      this.snackBar.open('Có lỗi khi tải danh sách hợp đồng', 'Đóng', {
        duration: 3000
      });
    } finally {
      this.loading = false;
    }
  }

  applyFilter(event: Event) {
    const filterValue = (event.target as HTMLInputElement).value.toLowerCase();
    this.filteredContracts = this.contracts.filter(contract =>
      contract.contractId.toLowerCase().includes(filterValue) ||
      contract.description.toLowerCase().includes(filterValue) ||
      contract.suppliers.some(s => s.name.toLowerCase().includes(filterValue))
    );
  }

  filterByStatus(event: MatSelectChange) {
    const status = event.value;
    if (status) {
      this.filteredContracts = this.contracts.filter(contract => contract.status === status);
    } else {
      this.filteredContracts = [...this.contracts];
    }
  }

  toggleContractDetails(contractId: string) {
    this.expandedContracts[contractId] = !this.expandedContracts[contractId];
  }


  trackByContractId(index: number, contract: Contract): string {
    return contract.contractId;
  }

  formatAmount(amount: number): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND'
    }).format(amount);
  }

  getSupplierStatus(supplier: any): string {
    return supplier.status || 'PENDING';
  }



  // Token information methods
  async loadTokenForContract(contractId: string): Promise<void> {
    try {
      const token = await firstValueFrom(this.contractTokenService.getToken(`token_${contractId}`));
      this.contractTokens[contractId] = token;
    } catch (error) {
      console.error(`Error loading token for contract ${contractId}:`, error);
    }
  }

  getTokenStatusForSupplier(contract: Contract, supplier: any): string {
    if (!this.contractTokens[contract.contractId]) {
      return 'Loading...';
    }

    const token = this.contractTokens[contract.contractId];
    if (token.owner === supplier.supplierId) {
      return 'Token Received';
    } else if (token.owner === contract.buyer) {
      return 'Token Pending';
    } else {
      return 'Token Transferred';
    }
  }

  getTokenStatusClass(supplierStatus: string): string {
    switch (supplierStatus) {
      case 'Token Received': return 'token-received';
      case 'Token Pending': return 'token-pending';
      case 'Token Transferred': return 'token-transferred';
      default: return 'token-loading';
    }
  }

  // Show token ownership details
  showTokenDetails(contract: Contract): void {
    const tokenId = `token_${contract.contractId}`;

    this.contractTokenService.getBalancesByToken(tokenId).subscribe({
      next: (balances) => {
        const dialogRef = this.dialog.open(TokenDetailsDialogComponent, {
          width: '600px',
          data: {
            tokenId: tokenId,
            contract: contract,
            balances: balances
          }
        });
      },
      error: (error) => {
        console.error('Error loading token balances:', error);
        this.snackBar.open('Không thể tải thông tin token', 'Đóng', {
          duration: 3000
        });
      }
    });
  }

  // Load token information when contract is expanded
  async onContractExpanded(contract: Contract): Promise<void> {
    if (!this.contractTokens[contract.contractId]) {
      await this.loadTokenForContract(contract.contractId);
    }
  }

}
