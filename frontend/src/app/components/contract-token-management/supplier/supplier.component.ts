import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ContractTokenService } from '../../../services/contract-token.service';
import { UserService } from '../../../services/user.service';
import { firstValueFrom, Observable, startWith, map } from 'rxjs';
declare var bootstrap: any; // Bootstrap modal

@Component({
  selector: 'app-supplier',
  templateUrl: './supplier.component.html',
  styleUrls: ['./supplier.component.css']
})
export class SupplierComponent implements OnInit {
  tokens: any[] = [];
  transferForm: FormGroup;
  quickTransferForm: FormGroup;
  loading = false;
  supplierId = ''; // Will be set from current user
  suppliers: any[] = []; // List of all suppliers for autocomplete
  filteredSuppliers: any[] = []; // Filtered suppliers for autocomplete
  balances: any[] = []; // User's token balances
  selectedSupplier: any = null; // Store selected supplier object for modal transfer
  selectedQuickSupplier: any = null; // Store selected supplier for quick transfer

  constructor(
    private fb: FormBuilder,
    private contractTokenService: ContractTokenService,
    private userService: UserService
  ) {
    this.transferForm = this.fb.group({
      tokenId: [''], // Will be set when opening modal
      to: ['', Validators.required],
      amount: [0, [Validators.required, Validators.min(0.01)]]
    });

    this.quickTransferForm = this.fb.group({
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

  selectTokenForTransfer(token: any): void {
    // Fill token ID into the quick transfer form on the right side
    this.quickTransferForm.patchValue({
      tokenId: token.id,
      to: '', // Reset recipient
      amount: '' // Reset amount
    });
    this.selectedQuickSupplier = null; // Reset selected supplier

    // Optional: Scroll to the transfer form or highlight it
    const transferForm = document.querySelector('.quick-transfer-form');
    if (transferForm) {
      transferForm.scrollIntoView({ behavior: 'smooth', block: 'center' });
      // Add a visual highlight effect
      transferForm.classList.add('highlight-form');
      setTimeout(() => {
        transferForm.classList.remove('highlight-form');
      }, 2000);
    }

    console.log(`Token ${token.symbol} selected for transfer`);
  }

  quickSettleToken(token: any): void {
    const balance = this.getTokenBalance(token.id);
    if (balance <= 0) {
      alert('You have no balance for this token to settle.');
      return;
    }

    // Show confirmation popup
    const confirmed = confirm(`Settle Token ${token.symbol}\n\nBalance to settle: ${balance.toLocaleString()} USD\n\n⚠️ This action cannot be undone!\n\nAre you sure you want to settle this token with the bank?`);

    if (confirmed) {
      this.loading = true;
      const settleData = {
        tokenId: token.id,
        supplierId: this.supplierId
      };

      this.contractTokenService.settleToken(settleData).subscribe({
        next: (response) => {
          console.log('Token settled with bank:', response);
          alert(`✅ Token ${token.symbol} settled successfully!\n\nBalance settled: ${balance.toLocaleString()} USD`);

          // Reset form selections and refresh data
          this.quickTransferForm.patchValue({ tokenId: '' });
          this.loadMyTokens();
          this.loadBalances();
        },
        error: (error) => {
          console.error('Error settling token with bank:', error);
          alert('❌ Error settling token with bank: ' + (error.message || 'Unknown error'));
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
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

          // Close modal
          const modal = bootstrap.Modal.getInstance(document.getElementById('transferModal'));
          modal.hide();

          // Reset form and selections
          this.transferForm.reset();
          this.selectedSupplier = null;

          // Refresh data
          this.loadBalances();
          this.loadMyTokens();
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


  onQuickTransfer(): void {
    if (this.quickTransferForm.valid && this.selectedQuickSupplier) {
      this.loading = true;
      const transferData = {
        tokenId: this.quickTransferForm.value.tokenId,
        from: this.supplierId,
        to: this.selectedQuickSupplier.id,
        amount: this.quickTransferForm.value.amount
      };

      this.contractTokenService.transferToken(transferData).subscribe({
        next: (response) => {
          console.log('Token transferred:', response);
          alert('Token transferred successfully!');
          this.quickTransferForm.reset();
          this.selectedQuickSupplier = null;
          this.loadBalances();
          this.loadMyTokens();
        },
        error: (error) => {
          console.error('Error transferring token:', error);
          let errorMessage = 'Error transferring token';
          if (error.error?.includes('Insufficient balance')) {
            errorMessage = 'Insufficient balance for this transfer';
          } else if (error.error?.includes('Token not found')) {
            errorMessage = 'Token not found';
          }
          alert(errorMessage);
          this.selectedQuickSupplier = null;
        },
        complete: () => {
          this.loading = false;
        }
      });
    }
  }

  selectQuickSupplier(event: any): void {
    const selectedSupplier = event.option.value;
    if (selectedSupplier) {
      const displayValue = this.displaySupplier(selectedSupplier);
      this.quickTransferForm.patchValue({
        to: displayValue
      });
      this.selectedQuickSupplier = selectedSupplier;
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
    console.log('DEBUG: Setting up autocomplete, suppliers loaded:', this.suppliers.length);

    // Set up autocomplete filtering for modal transfer form
    this.transferForm.get('to')?.valueChanges.pipe(
      startWith(''),
      map(value => this._filterSuppliers(value || ''))
    ).subscribe(filtered => {
      this.filteredSuppliers = filtered;
    });

    // Set up autocomplete filtering for quick transfer form
    this.quickTransferForm.get('to')?.valueChanges.pipe(
      startWith(''),
      map(value => this._filterSuppliers(value || ''))
    ).subscribe(filtered => {
      this.filteredSuppliers = filtered;
    });
  }

  private _filterSuppliers(value: string): any[] {
    const filterValue = value.toLowerCase();
    console.log('DEBUG: Filtering suppliers with value:', value, 'total suppliers:', this.suppliers.length);
    const filtered = this.suppliers.filter(supplier =>
      supplier.id.toLowerCase().includes(filterValue) ||
      (supplier.username && supplier.username.toLowerCase().includes(filterValue))
    );
    console.log('DEBUG: Filtered result:', filtered.length, 'suppliers');
    return filtered;
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

  get maxQuickTransferAmount(): number {
    const tokenId = this.quickTransferForm?.value?.tokenId;
    if (!tokenId) return 0;
    return this.getTokenBalance(tokenId);
  }
}
