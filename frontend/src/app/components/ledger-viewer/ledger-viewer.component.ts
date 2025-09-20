import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ContractService } from '../../services/contract.service';
import { MatSnackBar } from '@angular/material/snack-bar';

interface Transaction {
  id: string;
  type: string;
  timestamp: string;
  data: any;
}

interface Block {
  number: number;
  hash: string;
  previousHash: string;
  timestamp: string;
  transactions: Transaction[];
}

@Component({
  selector: 'app-ledger-viewer',
  templateUrl: './ledger-viewer.component.html',
  styleUrls: ['./ledger-viewer.component.css']
})
export class LedgerViewerComponent implements OnInit {
  contractId: string = '';
  transactions: Transaction[] = [];
  blocks: Block[] = [];
  loading = false;

  transactionColumns: string[] = ['id', 'type', 'timestamp', 'data'];
  blockColumns: string[] = ['number', 'hash', 'previousHash', 'timestamp', 'transactionCount'];

  constructor(
    private route: ActivatedRoute,
    private contractService: ContractService,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.contractId = params['id'];
      if (this.contractId) {
        this.loadLedger();
      }
    });
  }

  loadLedger() {
    this.loading = true;
    this.contractService.getContract(this.contractId).subscribe({
      next: (response: any) => {
        this.transactions = response.transactions || [];
        this.blocks = response.blocks || [];
        this.loading = false;
      },
      error: () => {
        this.snackBar.open('Error loading ledger data', 'Close', { duration: 3000 });
        this.loading = false;
      }
    });
  }

  getTransactionTypeColor(type: string): string {
    switch (type) {
      case 'CREATE': return 'primary';
      case 'APPROVE': return 'accent';
      case 'EXECUTE': return 'warn';
      default: return 'default';
    }
  }
}