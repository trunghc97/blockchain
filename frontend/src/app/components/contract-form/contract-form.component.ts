import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { firstValueFrom } from 'rxjs';
import { ContractTokenService } from '../../services/contract-token.service';
import { UserService } from '../../services/user.service';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-contract-form',
  templateUrl: './contract-form.component.html',
  styleUrls: ['./contract-form.component.css']
})
export class ContractFormComponent implements OnInit {
  contractForm: FormGroup;
  loading = false;
  availableSuppliers: any[] = [];
  totalAmount = 0;

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService,
    private userService: UserService,
    private router: Router,
    private snackBar: MatSnackBar
  ) {
    this.contractForm = this.fb.group({
      description: ['', Validators.required],
      suppliers: this.fb.array([this.createSupplierGroup()])
    });

    // Calculate total amount when suppliers change
    this.contractForm.get('suppliers')?.valueChanges.subscribe(suppliers => {
      this.calculateTotalAmount();
    });
  }

  ngOnInit() {
    this.loadSuppliers();
  }

  get suppliers() {
    return this.contractForm.get('suppliers') as FormArray;
  }

  createSupplierGroup(): FormGroup {
    const group = this.fb.group({
      supplierId: ['', Validators.required],
      name: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]]
    });

    // Auto-fill name when supplierId changes
    group.get('supplierId')?.valueChanges.subscribe(supplierId => {
      if (supplierId) {
        const selectedSupplier = this.availableSuppliers.find(s => s.id === supplierId);
        if (selectedSupplier) {
          group.patchValue({ name: selectedSupplier.username }, { emitEvent: false });
        }
      }
    });

    return group;
  }

  async loadSuppliers() {
    try {
      this.availableSuppliers = await firstValueFrom(this.contractTokenService.getSuppliers());
      console.log('Available suppliers:', this.availableSuppliers);
    } catch (error) {
      console.error('Error loading suppliers:', error);
      this.snackBar.open('Không thể tải danh sách nhà cung cấp', 'Đóng', { duration: 3000 });
    }
  }

  addSupplier() {
    this.suppliers.push(this.createSupplierGroup());
  }

  removeSupplier(index: number) {
    if (this.suppliers.length > 1) {
      this.suppliers.removeAt(index);
      this.calculateTotalAmount();
    }
  }


  async onSubmit() {
    if (this.contractForm.valid) {
      this.loading = true;
      try {
        const formValue = this.contractForm.value;

        // Transform suppliers to include status field
        const suppliers = formValue.suppliers.map((supplier: any) => ({
          supplierId: supplier.supplierId,
          name: supplier.name,
          amount: supplier.amount,
          status: 'PENDING'
        }));

        const contractData = {
          description: formValue.description,
          suppliers: suppliers,
          // anchorId, bankId will be auto-filled by backend
          // approvers removed
        };

        const result = await firstValueFrom(this.contractTokenService.createContract(contractData));
        this.snackBar.open('Tạo hợp đồng thành công!', 'Đóng', { duration: 3000 });
        this.contractForm.reset();
        this.router.navigate(['/contracts']);
      } catch (error: any) {
        console.error('Error creating contract:', error);
        this.snackBar.open(`Có lỗi xảy ra khi tạo hợp đồng: ${error.message || error}`, 'Đóng', {
          duration: 3000
        });
      } finally {
        this.loading = false;
      }
    } else {
      this.snackBar.open('Vui lòng điền đầy đủ thông tin', 'Đóng', { duration: 3000 });
    }
  }

  calculateTotalAmount(): void {
    const suppliers = this.suppliers.value;
    this.totalAmount = suppliers.reduce((total: number, supplier: any) => {
      return total + (supplier.amount || 0);
    }, 0);
  }
}