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
  
  transactionColumns: string[] = ['id', 'type', 'approverID', 'status', 'timestamp', 'blockNumber', 'payload'];
  blockColumns: string[] = ['blockNumber', 'timestamp', 'hash', 'previousHash', 'merkleRoot', 'txIds'];
  
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
        // Transform new API format to old format for frontend compatibility
        const transformedData = this.contractService.transformLedgerData(data);
        this.transactions.data = transformedData.transactions || [];
        this.blocks.data = transformedData.blocks || [];
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

  hasPayload(payload: any): boolean {
    return payload && typeof payload === 'object' && Object.keys(payload).length > 0;
  }

  formatPayload(payload: any): string {
    return JSON.stringify(payload, null, 2);
  }

  getTransactionTypeColor(type: string): string {
    switch (type.toUpperCase()) {
      case 'CREATE':
        return 'primary';
      case 'APPROVE':
        return 'accent';
      case 'REJECT':
        return 'warn';
      case 'CURRENT_STATE':
        return 'primary';
      default:
        return '';
    }
  }
}
