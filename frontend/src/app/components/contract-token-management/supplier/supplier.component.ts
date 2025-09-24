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
  selectedSupplier: any = null; // Store selected supplier object

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
    if (this.transferForm.valid && this.selectedSupplier) {
      this.loading = true;
      const transferData = {
        tokenId: this.transferForm.value.tokenId,
        from: this.supplierId,
        to: this.selectedSupplier.id, // Use selected supplier ID
        amount: this.transferForm.value.amount
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token transferred:', response);
          alert('Token transferred successfully!');
          this.transferForm.reset();
          this.selectedSupplier = null; // Reset selected supplier
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
          this.selectedSupplier = null; // Reset on error too
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  onPayBack(token: any): void {
    const balance = this.getTokenBalance(token.id);
    if (balance <= 0) {
      alert('You have no balance for this token to settle.');
      return;
    }

    if (confirm(`Are you sure you want to settle this token with the bank?\nAmount to settle: ${balance}\nThis will remove all your balance for this token.`)) {
      this.loading = true;
      const settleData = {
        tokenId: token.id,
        supplierId: this.supplierId
      };

      this.contractTokenService.settleToken(settleData).subscribe({
        next: (response) => {
          console.log('Token settled with bank:', response);
          alert('Token settled with bank successfully!');
          this.loadMyTokens();
          this.loadBalances();
        },
        error: (error) => {
          console.error('Error settling token with bank:', error);
          alert('Error settling token with bank: ' + error.message);
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
      (supplier.username && supplier.username.toLowerCase().includes(filterValue))
    );
  }

  displaySupplier(supplier: any): string {
    if (!supplier) return '';
    const name = supplier.username || 'Unknown';
    return `${supplier.id} (${name})`;
  }

  selectSupplier(event: any): void {
    const selectedSupplier = event.option.value;
    if (selectedSupplier) {
      // Set the form value to the full display string for better UX
      // But we'll need to extract the ID when submitting
      const displayValue = this.displaySupplier(selectedSupplier);
      this.transferForm.patchValue({
        to: displayValue
      });
      // Store the actual supplier object for submission
      this.selectedSupplier = selectedSupplier;
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
