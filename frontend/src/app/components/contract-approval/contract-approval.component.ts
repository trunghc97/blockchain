import { Component, OnInit } from '@angular/core';
import { ContractService } from '../../services/contract.service';
import { UserService } from '../../services/user.service';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Contract } from '../../models/contract.model';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-contract-approval',
  templateUrl: './contract-approval.component.html',
  styleUrls: ['./contract-approval.component.css']
})
export class ContractApprovalComponent implements OnInit {
  contracts: Contract[] = [];
  loading = false;
  approving: { [key: string]: boolean } = {};
  displayedColumns: string[] = ['id', 'description', 'totalAmount', 'status', 'actions'];

  constructor(
    private contractService: ContractService,
    private userService: UserService,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    this.loadContracts();
  }

  loadContracts() {
    this.loading = true;
    this.contractService.getContracts().subscribe({
      next: (contracts: Contract[]) => {
        const currentUser = this.userService.getCurrentUser();
        if (!currentUser) {
          this.snackBar.open('User not authenticated', 'Close', { duration: 3000 });
          return;
        }

        this.contracts = contracts.filter((contract: Contract) => {
          // For suppliers, only show contracts where they are listed
          if (currentUser.role === 'SUPPLIER') {
            return contract.suppliers.some(s =>
              s.supplierId === currentUser.id
            );
          }
          // For other roles, show all contracts
          return true;
        });
        this.loading = false;
      },
      error: () => {
        this.snackBar.open('Error loading contracts', 'Close', { duration: 3000 });
        this.loading = false;
      }
    });
  }

  approveContract(contract: Contract) {
    if (!contract.contractId) return;
    
    this.approving[contract.contractId] = true;
    const currentUser = this.userService.getCurrentUser();
    if (!currentUser) return;

    this.contractService.approveContract(contract.contractId).subscribe({
      next: () => {
        this.snackBar.open('Contract approved successfully', 'Close', { duration: 3000 });
        this.loadContracts();
      },
      error: () => {
        this.snackBar.open('Error approving contract', 'Close', { duration: 3000 });
      },
      complete: () => {
        if (contract.contractId) {
          this.approving[contract.contractId] = false;
        }
      }
    });
  }

  formatAmount(amount: number): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND'
    }).format(amount);
  }
}