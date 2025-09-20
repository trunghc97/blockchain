import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { firstValueFrom } from 'rxjs';
import { Contract, SupplierAmount } from '../../models/contract.model';
import { ContractService } from '../../services/contract.service';
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
  selectedFile: File | null = null;
  supplierList: User[] = [];

  constructor(
    private fb: FormBuilder,
    private contractService: ContractService,
    private userService: UserService,
    private router: Router,
    private snackBar: MatSnackBar
  ) {
    this.contractForm = this.fb.group({
      description: ['', Validators.required],
      suppliers: this.fb.array([])
    });
  }

  ngOnInit() {
    this.loadSuppliers();
    this.addSupplier();
  }

  get suppliers() {
    return this.contractForm.get('suppliers') as FormArray;
  }

  async loadSuppliers() {
    try {
      this.supplierList = await firstValueFrom(this.userService.getSuppliers());
    } catch (error) {
      console.error('Error loading suppliers:', error);
      this.snackBar.open('Có lỗi khi tải danh sách supplier', 'Đóng', {
        duration: 3000
      });
    }
  }

  addSupplier() {
    const supplierGroup = this.fb.group({
      supplierId: ['', Validators.required],
      name: [{ value: '', disabled: true }, Validators.required],
      amount: [0, [Validators.required, Validators.min(1)]],
      status: ['PENDING']
    });

    this.suppliers.push(supplierGroup);
  }

  removeSupplier(index: number) {
    if (this.suppliers.length > 1) {
      this.suppliers.removeAt(index);
    }
  }

  onFileSelected(event: any) {
    const file = event.target.files[0];
    if (file) {
      this.selectedFile = file;
    }
  }

  removeFile() {
    this.selectedFile = null;
  }

  async onSubmit() {
    this.markAllAsTouched(this.contractForm);
    if (this.contractForm.invalid) {
      this.snackBar.open('Vui lòng điền đầy đủ thông tin và đảm bảo số tiền lớn hơn 0', 'Đóng', {
        duration: 3000
      });
      return;
    }

    this.loading = true;

    try {
      const currentUser = await firstValueFrom(this.userService.getCurrentUser());

      if (!currentUser) {
        throw new Error('User not authenticated');
      }

      const supplierAmounts: SupplierAmount[] = this.contractForm.value.suppliers.map((s: any) => ({
        supplierId: s.supplierId,
        name: this.supplierList.find((sup: User) => sup.id === s.supplierId)?.username || '',
        amount: s.amount,
        status: 'PENDING'
      }));

      const contractData: Partial<Contract> = {
        description: this.contractForm.value.description,
        buyer: currentUser.id,
        suppliers: supplierAmounts,
        status: 'PENDING',
        totalAmount: this.calculateTotalAmount(),
        createdAt: new Date(),
        updatedAt: new Date()
      };

      if (this.selectedFile) {
        const formData = new FormData();
        formData.append('file', this.selectedFile);
        formData.append('contract', JSON.stringify(contractData));

        await firstValueFrom(this.contractService.createContractWithFile(formData));
      } else {
        await firstValueFrom(this.contractService.createContract(contractData));
      }

      this.snackBar.open('Tạo hợp đồng thành công', 'Đóng', {
        duration: 3000
      });
      this.router.navigate(['/contracts']);
    } catch (error: any) {
      console.error('Error creating contract:', error);
      this.snackBar.open(`Có lỗi xảy ra khi tạo hợp đồng: ${error.message || error}`, 'Đóng', {
        duration: 3000
      });
    } finally {
      this.loading = false;
    }
  }

  private calculateTotalAmount(): number {
    return this.suppliers.controls.reduce((total, control) => {
      return total + (control.get('amount')?.value || 0);
    }, 0);
  }

  onSupplierSelected(event: any, index: number) {
    const supplierId = event.value;
    const supplier = this.supplierList.find(s => s.id === supplierId);
    if (supplier) {
      const supplierGroup = this.suppliers.at(index);
      supplierGroup.patchValue({
        name: supplier.username
      });
    }
  }

  isFormValid(): boolean {
    return this.contractForm.valid && !this.loading;
  }

  markAllAsTouched(formGroup: FormGroup | FormArray) {
    Object.values(formGroup.controls).forEach(control => {
      control.markAsTouched();
      if (control instanceof FormGroup || control instanceof FormArray) {
        this.markAllAsTouched(control);
      }
    });
  }
}