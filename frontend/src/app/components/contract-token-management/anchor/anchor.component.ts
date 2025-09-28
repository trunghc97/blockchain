import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';
import { ContractTokenService } from '../../../services/contract-token.service';

@Component({
  selector: 'app-anchor',
  templateUrl: './anchor.component.html',
  styleUrls: ['./anchor.component.css']
})
export class AnchorComponent {
  contractForm: FormGroup;
  contracts: any[] = [];
  loading = false;

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService
  ) {
    this.contractForm = this.fb.group({
      id: [''],
      anchorId: ['', Validators.required],
      supplierId: ['', Validators.required],
      bankId: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]],
      approvers: this.fb.array([this.fb.control('', Validators.required)])
    });
  }

  get approvers(): FormArray {
    return this.contractForm.get('approvers') as FormArray;
  }

  addApprover(): void {
    this.approvers.push(this.fb.control('', Validators.required));
  }

  removeApprover(index: number): void {
    if (this.approvers.length > 1) {
      this.approvers.removeAt(index);
    }
  }

  onSubmit(): void {
    if (this.contractForm.valid) {
      this.loading = true;
      const contractData = {
        ...this.contractForm.value,
        approvers: this.approvers.value.filter((approver: string) => approver.trim() !== '')
      };

      this.contractTokenService.createContract(contractData).subscribe({
        next: (response) => {
          console.log('Contract created:', response);
          alert('Contract created successfully!');
          this.contractForm.reset();
          this.loadContracts();
        },
        error: (error) => {
          console.error('Error creating contract:', error);
          alert('Error creating contract: ' + error.message);
        },
        complete: () => {
          this.loading = false;
        }
      });
    } else {
      alert('Please fill in all required fields correctly.');
    }
  }

  loadContracts(): void {
    // Load all contracts - anchor can see all contracts in the system
    this.contractTokenService.getContracts().subscribe({
      next: (allContracts: any[]) => {
        this.contracts = allContracts;
        console.log('Loaded contracts for anchor:', this.contracts);
      },
      error: (error: any) => {
        console.error('Error loading contracts:', error);
        this.contracts = [];
      }
    });
  }

  ngOnInit(): void {
    this.loadContracts();
  }

  getStatusBadgeClass(status: string): string {
    switch (status) {
      case 'BANK_APPROVED':
        return 'bg-info';
      case 'EXECUTED':
        return 'bg-success';
      case 'PENDING_BANK_APPROVAL':
      default:
        return 'bg-warning';
    }
  }

  getStatusText(status: string): string {
    switch (status) {
      case 'BANK_APPROVED':
        return 'Bank Approved';
      case 'EXECUTED':
        return 'Executed';
      case 'PENDING_BANK_APPROVAL':
        return 'Pending Bank Approval';
      default:
        return status || 'Unknown';
    }
  }
}
