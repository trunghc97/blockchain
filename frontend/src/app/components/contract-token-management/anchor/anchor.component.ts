import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, FormArray } from '@angular/forms';
import { ContractTokenService } from '../../../services/contract-token.service';
import { ContractService } from '../../../services/contract.service';
import { firstValueFrom } from 'rxjs';

@Component({
  selector: 'app-anchor',
  templateUrl: './anchor.component.html',
  styleUrls: ['./anchor.component.css']
})
export class AnchorComponent implements OnInit {
  contractForm: FormGroup;
  contracts: any[] = [];
  loading = false;

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService,
    private contractService: ContractService
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

  async loadContracts(): Promise<void> {
    try {
      this.loading = true;
      // Use new API - backend will filter contracts based on authenticated user role
      this.contracts = await firstValueFrom(this.contractService.getContracts());
      console.log('Loaded contracts for anchor:', this.contracts);
    } catch (error) {
      console.error('Error loading contracts:', error);
      alert('Error loading contracts: ' + (error as any).message);
    } finally {
      this.loading = false;
    }
  }

  ngOnInit(): void {
    this.loadContracts();
  }
}
