import { Component, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { firstValueFrom } from 'rxjs';
import { Contract } from '../../models/contract.model';
import { ContractService } from '../../services/contract.service';
import { UserService } from '../../services/user.service';
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
  displayedColumns: string[] = ['contractId', 'description', 'totalAmount', 'status', 'actions'];
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
    this.loading = true;
    try {
      this.currentUser = await firstValueFrom(this.userService.getCurrentUser());
      console.log('Current user:', this.currentUser);

      await this.loadContracts();
    } catch (error) {
      console.error('Error loading current user:', error);
    } finally {
      this.loading = false;
    }
  }

  async loadContracts() {
    try {
      const allContracts = await firstValueFrom(this.contractService.getContracts());

      // Lọc contracts có suppliers mà user hiện tại có thể approve
      this.contracts = allContracts.filter(contract =>
        contract.suppliers.some(s =>
          s.supplierId === this.currentUser?.id && s.status === 'PENDING'
        )
      );

      console.log('Filtered contracts for approval:', this.contracts);

      // Khởi tạo trạng thái approving cho mỗi contract
      this.contracts.forEach(contract => {
        this.approving[contract.contractId] = false;
      });
    } catch (error) {
      console.error('Error loading contracts:', error);
      this.snackBar.open('Có lỗi khi tải danh sách hợp đồng', 'Đóng', {
        duration: 3000
      });
    }
  }

  async approveContract(contract: Contract) {
    this.approving[contract.contractId] = true;
    try {
      await firstValueFrom(this.contractService.approveContract(contract.contractId));
      this.snackBar.open('Đã duyệt hợp đồng thành công', 'Đóng', {
        duration: 3000
      });
      await this.loadContracts();
    } catch (error: any) {
      console.error('Error approving contract:', error);
      const errorMessage = error.error?.message || 'Có lỗi khi duyệt hợp đồng';
      this.snackBar.open(errorMessage, 'Đóng', {
        duration: 3000
      });
    } finally {
      this.approving[contract.contractId] = false;
    }
  }

  async rejectContract(contract: Contract) {
    this.approving[contract.contractId] = true;
    try {
      await firstValueFrom(this.contractService.rejectContract(contract.contractId, 'Rejected by supplier'));
      this.snackBar.open('Đã từ chối hợp đồng', 'Đóng', {
        duration: 3000
      });
      await this.loadContracts();
    } catch (error: any) {
      console.error('Error rejecting contract:', error);
      const errorMessage = error.error?.message || 'Có lỗi khi từ chối hợp đồng';
      this.snackBar.open(errorMessage, 'Đóng', {
        duration: 3000
      });
    } finally {
      this.approving[contract.contractId] = false;
    }
  }

  formatAmount(amount: number): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND'
    }).format(amount);
  }

  // Get the supplier info for current user in this contract
  getCurrentUserSupplier(contract: Contract): any {
    return contract.suppliers.find(s => s.supplierId === this.currentUser?.id);
  }

  // Check if current user can approve this contract
  canApproveContract(contract: Contract): boolean {
    const supplier = this.getCurrentUserSupplier(contract);
    return !!(supplier && supplier.status === 'PENDING');
  }
}