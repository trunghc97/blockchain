import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, FormArray, Validators } from '@angular/forms';
import { ContractService } from '../../services/contract.service';
import { UserService } from '../../services/user.service';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Contract, SupplierAmount } from '../../models/contract.model';
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
    this.addSupplier(); // Add first supplier by default
  }

  get suppliers() {
    return this.contractForm.get('suppliers') as FormArray;
  }

  loadSuppliers() {
    this.userService.getSuppliers().subscribe(
      (suppliers: User[]) => this.supplierList = suppliers,
      (error: Error) => this.snackBar.open('Error loading suppliers', 'Close', { duration: 3000 })
    );
  }

  addSupplier() {
    const supplierForm = this.fb.group({
      supplierId: ['', Validators.required],
      amount: ['', [Validators.required, Validators.min(0)]]
    });

    this.suppliers.push(supplierForm);
  }

  removeSupplier(index: number) {
    this.suppliers.removeAt(index);
  }

  onFileSelected(event: Event) {
    const element = event.target as HTMLInputElement;
    const file = element.files?.[0];
    if (file) {
      this.selectedFile = file;
    }
  }

  removeFile() {
    this.selectedFile = null;
  }

  onSubmit() {
    if (this.contractForm.valid) {
      this.loading = true;

      const formData = new FormData();
      if (this.selectedFile) {
        formData.append('file', this.selectedFile);
      }

      const contract: Contract = {
        description: this.contractForm.value.description,
        suppliers: this.contractForm.value.suppliers as SupplierAmount[]
      };

      formData.append('contract', JSON.stringify(contract));

      this.contractService.createContract(formData).subscribe({
        next: (response: Contract) => {
          this.snackBar.open('Contract created successfully', 'Close', { duration: 3000 });
          this.router.navigate(['/contracts/status']);
        },
        error: (error: Error) => {
          console.error('Error creating contract:', error);
          this.snackBar.open('Error creating contract', 'Close', { duration: 3000 });
          this.loading = false;
        },
        complete: () => {
          this.loading = false;
        }
      });
    } else {
      this.snackBar.open('Please fill in all required fields', 'Close', { duration: 3000 });
    }
  }

  formatAmount(amount: number): string {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND'
    }).format(amount);
  }
}