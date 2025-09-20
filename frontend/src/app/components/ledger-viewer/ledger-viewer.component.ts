import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ContractService } from '../../services/contract.service';

@Component({
  selector: 'app-ledger-viewer',
  templateUrl: './ledger-viewer.component.html',
  styleUrls: ['./ledger-viewer.component.css']
})
export class LedgerViewerComponent implements OnInit {
  contractId: string = '';
  loading = false;
  
  transactionColumns: string[] = ['id', 'type', 'approverID', 'status', 'timestamp'];
  blockColumns: string[] = ['blockNumber', 'timestamp', 'hash', 'previousHash', 'txIds'];
  
  transactions = new MatTableDataSource<any>([]);
  blocks = new MatTableDataSource<any>([]);

  @ViewChild('transactionPaginator') transactionPaginator!: MatPaginator;
  @ViewChild('blockPaginator') blockPaginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private route: ActivatedRoute,
    private contractService: ContractService,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      this.contractId = params['contractId'];
      if (this.contractId) {
        this.loadLedgerData();
      }
    });
  }

  ngAfterViewInit() {
    this.transactions.paginator = this.transactionPaginator;
    this.blocks.paginator = this.blockPaginator;
    this.transactions.sort = this.sort;
    this.blocks.sort = this.sort;
  }

  loadLedgerData() {
    this.loading = true;
    this.contractService.getLedgerData(this.contractId).subscribe({
      next: (data) => {
        this.transactions.data = data.transactions;
        this.blocks.data = data.blocks;
        this.loading = false;
      },
      error: (error) => {
        console.error('Error loading ledger data:', error);
        this.loading = false;
      }
    });
  }

  copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      this.snackBar.open('Đã sao chép vào clipboard', 'Đóng', {
        duration: 2000,
        horizontalPosition: 'center',
        verticalPosition: 'bottom'
      });
    });
  }

  getTransactionTypeColor(type: string): string {
    switch (type.toUpperCase()) {
      case 'CREATE':
        return 'primary';
      case 'APPROVE':
        return 'accent';
      case 'REJECT':
        return 'warn';
      default:
        return '';
    }
  }
}