import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { UserService } from '../../services/user.service';
import { TransferService } from '../../services/transfer.service';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-transfer-form',
  templateUrl: './transfer-form.component.html',
  styleUrls: ['./transfer-form.component.css']
})
export class TransferFormComponent implements OnInit {
  transferForm: FormGroup;
  users: User[] = [];
  selectedApprovers: User[] = [];

  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private transferService: TransferService,
    private snackBar: MatSnackBar
  ) {
    this.transferForm = this.fb.group({
      toAccount: ['', [Validators.required]],
      amount: ['', [Validators.required, Validators.min(0)]],
      description: ['', [Validators.required]],
      approvers: [[], [Validators.required, Validators.minLength(1)]]
    });
  }

  ngOnInit(): void {
    this.loadUsers();
  }

  loadUsers(): void {
    this.userService.getUsers().subscribe(users => {
      this.users = users;
    });
  }

  onSubmit(): void {
    if (this.transferForm.valid) {
      const formValue = this.transferForm.value;
      const request = {
        fromUser: localStorage.getItem('currentUserId') || '',
        toAccount: formValue.toAccount,
        amount: formValue.amount,
        description: formValue.description,
        approvers: formValue.approvers.map((user: User) => user.id)
      };

      this.transferService.createTransfer(request).subscribe(
        () => {
          this.snackBar.open('Giao dịch đã được tạo thành công', 'Đóng', {
            duration: 3000
          });
          this.transferForm.reset();
        },
        error => {
          this.snackBar.open('Có lỗi xảy ra khi tạo giao dịch', 'Đóng', {
            duration: 3000
          });
        }
      );
    }
  }
}