import { Component, OnInit } from '@angular/core';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatSelectChange } from '@angular/material/select';
import { MatSnackBar } from '@angular/material/snack-bar';
import { firstValueFrom } from 'rxjs';
import { Contract } from '../../models/contract.model';
import { ContractService } from '../../services/contract.service';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user.model';

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
  approvingSuppliers: { [key: string]: boolean } = {};
  currentUser: User | null = null;

  constructor(
    private contractService: ContractService,
    private userService: UserService,
    private snackBar: MatSnackBar
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

  async approveSupplier(contract: Contract, supplier: any) {
    this.approvingSuppliers[supplier.supplierId] = true;
    try {
      // Use contract.contractId for the API call
      await firstValueFrom(this.contractService.approveContract(contract.contractId));
      await this.loadContracts(); // Reload to get updated data

      this.snackBar.open(`Đã duyệt ${supplier.name} thành công`, 'Đóng', {
        duration: 3000
      });
    } catch (error) {
      console.error('Error approving supplier:', error);
      this.snackBar.open('Có lỗi khi duyệt nhà cung cấp', 'Đóng', {
        duration: 3000
      });
    } finally {
      this.approvingSuppliers[supplier.supplierId] = false;
    }
  }

  async rejectSupplier(contract: Contract, supplier: any) {
    this.approvingSuppliers[supplier.supplierId] = true;
    try {
      // Use contract.contractId for the API call
      await firstValueFrom(this.contractService.rejectContract(contract.contractId, 'Rejected by supplier'));
      await this.loadContracts(); // Reload to get updated data

      this.snackBar.open(`Đã từ chối ${supplier.name}`, 'Đóng', {
        duration: 3000
      });
    } catch (error) {
      console.error('Error rejecting supplier:', error);
      this.snackBar.open('Có lỗi khi từ chối nhà cung cấp', 'Đóng', {
        duration: 3000
      });
    } finally {
      this.approvingSuppliers[supplier.supplierId] = false;
    }
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

  // Check if current user can approve this supplier
  canApproveSupplier(supplier: any): boolean {
    return !!(this.currentUser && this.currentUser.id === supplier.supplierId);
  }
}