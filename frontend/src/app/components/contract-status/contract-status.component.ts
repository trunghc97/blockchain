import { Component, OnInit } from '@angular/core';
import { ContractService } from '../../services/contract.service';
import { UserService } from '../../services/user.service';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Contract } from '../../models/contract.model';

@Component({
  selector: 'app-contract-status',
  templateUrl: './contract-status.component.html',
  styleUrls: ['./contract-status.component.css']
})
export class ContractStatusComponent implements OnInit {
  contracts: Contract[] = [];
  loading = false;
  displayedColumns: string[] = ['id', 'description', 'suppliers', 'totalAmount', 'status', 'actions'];

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
        this.contracts = contracts;
        this.loading = false;
      },
      error: () => {
        this.snackBar.open('Error loading contracts', 'Close', { duration: 3000 });
        this.loading = false;
      }
    });
  }

  getStatusColor(status: string): string {
    switch (status) {
      case 'PENDING': return 'warn';
      case 'APPROVED': return 'primary';
      case 'EXECUTED': return 'accent';
      default: return 'default';
    }
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
}