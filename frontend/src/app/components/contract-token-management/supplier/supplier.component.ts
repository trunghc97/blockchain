import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ContractTokenService } from '../../../services/contract-token.service';
import { UserService } from '../../../services/user.service';
import { firstValueFrom, Observable, startWith, map } from 'rxjs';

@Component({
  selector: 'app-supplier',
  templateUrl: './supplier.component.html',
  styleUrls: ['./supplier.component.css']
})
export class SupplierComponent implements OnInit {
  tokens: any[] = [];
  transferForm: FormGroup;
  loading = false;
  supplierId = ''; // Will be set from current user
  suppliers: any[] = []; // List of all suppliers for autocomplete
  filteredSuppliers: any[] = []; // Filtered suppliers for autocomplete
  balances: any[] = []; // User's token balances

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService,
    private userService: UserService
  ) {
    this.transferForm = this.fb.group({
      tokenId: ['', Validators.required],
      to: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]]
    });
  }

  async ngOnInit(): Promise<void> {
    try {
      const currentUser = await firstValueFrom(this.userService.getCurrentUser());
      this.supplierId = currentUser.id;
      console.log('Current supplier:', this.supplierId);
      this.loadSuppliers();
      this.loadMyTokens();
      this.loadBalances();
      this.setupAutocomplete();
    } catch (error) {
      console.error('Error getting current user:', error);
    }
  }

  loadMyTokens(): void {
    this.loading = true;
    // Load all tokens since ownership no longer determines who can transfer balances
    // The transfer validation will check if user has sufficient balance for the token
    this.contractTokenService.getAllTokens().subscribe({
      next: (allTokens) => {
        this.tokens = allTokens; // Show all available tokens
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading tokens:', error);
        alert('Error loading tokens: ' + error.message);
        this.loading = false;
      }
    });
  }

  onTransfer(): void {
    if (this.transferForm.valid) {
      this.loading = true;
      const transferData = {
        tokenId: this.transferForm.value.tokenId,
        from: this.supplierId,
        to: this.transferForm.value.to,
        amount: this.transferForm.value.amount
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token transferred:', response);
          alert('Token transferred successfully!');
          this.transferForm.reset();
          this.loadBalances();
        },
        error: (error) => {
          console.error('Error transferring token:', error);

          // Provide more user-friendly error messages
          let errorMessage = 'Error transferring token';
          if (error.error?.includes('Insufficient balance')) {
            errorMessage = 'Insufficient balance for this transfer';
          } else if (error.error?.includes('Sender is not the current owner')) {
            errorMessage = 'You are not the current owner of this token';
          } else if (error.error?.includes('Token not found')) {
            errorMessage = 'Token not found';
          } else if (error.status === 400) {
            errorMessage = 'Invalid transfer request. Please check your input.';
          }

          alert(errorMessage);
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  onPayBack(token: any): void {
    if (confirm(`Are you sure you want to return this token to the bank?`)) {
      this.loading = true;
      const transferData = {
        tokenId: token.id,
        from: this.supplierId,
        to: token.issuer, // Bank ID
        amount: token.total
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token returned to bank:', response);
          alert('Token returned to bank successfully!');
          this.loadMyTokens();
        },
        error: (error) => {
          console.error('Error returning token to bank:', error);
          alert('Error returning token to bank: ' + error.message);
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  loadSuppliers(): void {
    this.userService.getSuppliers().subscribe({
      next: (suppliers) => {
        // Filter out current user from the list
        this.suppliers = suppliers.filter(supplier => supplier.id !== this.supplierId);
        console.log('Loaded suppliers for autocomplete:', this.suppliers);
      },
      error: (error) => {
        console.error('Error loading suppliers:', error);
      }
    });
  }

  loadBalances(): void {
    this.contractTokenService.getBalancesByAccount(this.supplierId).subscribe({
      next: (balances) => {
        this.balances = balances;
        console.log('Loaded balances:', this.balances);
      },
      error: (error) => {
        console.error('Error loading balances:', error);
        this.balances = [];
      }
    });
  }

  setupAutocomplete(): void {
    // Set up autocomplete filtering
    this.transferForm.get('to')?.valueChanges.pipe(
      startWith(''),
      map(value => this._filterSuppliers(value || ''))
    ).subscribe(filtered => {
      this.filteredSuppliers = filtered;
    });
  }

  private _filterSuppliers(value: string): any[] {
    const filterValue = value.toLowerCase();
    return this.suppliers.filter(supplier =>
      supplier.id.toLowerCase().includes(filterValue) ||
      supplier.username.toLowerCase().includes(filterValue) ||
      (supplier.name && supplier.name.toLowerCase().includes(filterValue))
    );
  }

  displaySupplier(supplier: any): string {
    return supplier ? supplier.id : '';
  }

  selectSupplier(event: any): void {
    const selectedSupplier = event.option.value;
    if (selectedSupplier) {
      this.transferForm.patchValue({
        to: selectedSupplier.id
      });
    }
  }

  get availableTokens(): any[] {
    // Return all tokens since any token can be transferred if user has balance
    return this.tokens;
  }

  getTokenBalance(tokenId: string): number {
    const balanceEntry = this.balances.find(b => b.tokenId === tokenId);
    return balanceEntry ? balanceEntry.balance : 0;
  }
}
