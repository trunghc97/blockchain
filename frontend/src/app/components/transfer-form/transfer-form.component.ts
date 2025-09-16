import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { TransferService } from '../../services/transfer.service';
import { v4 as uuidv4 } from 'uuid';

@Component({
  selector: 'app-transfer-form',
  templateUrl: './transfer-form.component.html',
  styleUrls: ['./transfer-form.component.css']
})
export class TransferFormComponent {
  transferForm: FormGroup;

  constructor(
    private fb: FormBuilder,
    private transferService: TransferService
  ) {
    this.transferForm = this.fb.group({
      fromAccount: ['', Validators.required],
      toAccount: ['', Validators.required],
      amount: ['', [Validators.required, Validators.min(0)]]
    });
  }

  onSubmit() {
    if (this.transferForm.valid) {
      const request = {
        transactionId: uuidv4(),
        ...this.transferForm.value
      };

      this.transferService.createTransfer(request).subscribe({
        next: (response) => {
          console.log('Transfer created:', response);
          this.transferForm.reset();
        },
        error: (error) => {
          console.error('Error creating transfer:', error);
        }
      });
    }
  }
}
