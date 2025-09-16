import { Component, OnInit } from '@angular/core';
import { TransferService } from '../../services/transfer.service';
import { UserService } from '../../services/user.service';
import { TransferStatus } from '../../models/transfer.model';
import { User } from '../../models/user.model';

@Component({
  selector: 'app-status-list',
  templateUrl: './status-list.component.html',
  styleUrls: ['./status-list.component.css']
})
export class StatusListComponent implements OnInit {
  transfers: TransferStatus[] = [];
  users: { [key: string]: User } = {};
  displayedColumns: string[] = [
    'reqId',
    'fromUser',
    'toAccount',
    'amount',
    'description',
    'status',
    'approvers'
  ];

  constructor(
    private transferService: TransferService,
    private userService: UserService
  ) {}

  ngOnInit(): void {
    this.loadUsers();
    this.loadTransfers();
  }

  loadUsers(): void {
    this.userService.getUsers().subscribe(users => {
      users.forEach(user => {
        this.users[user.id] = user;
      });
    });
  }

  loadTransfers(): void {
    this.transferService.getTransferList().subscribe(transfers => {
      this.transfers = transfers;
    });
  }

  getUserName(userId: string): string {
    return this.users[userId]?.name || userId;
  }

  getStatusLabel(status: string): string {
    switch (status) {
      case 'PENDING':
        return 'Chờ duyệt';
      case 'PARTIALLY_APPROVED':
        return 'Đã duyệt một phần';
      case 'EXECUTED':
        return 'Đã thực hiện';
      default:
        return status;
    }
  }

  getStatusColor(status: string): string {
    switch (status) {
      case 'PENDING':
        return 'primary';
      case 'PARTIALLY_APPROVED':
        return 'accent';
      case 'EXECUTED':
        return 'success';
      default:
        return '';
    }
  }
}
