import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TransferRequest, WorldState } from '../models/transfer.model';

@Injectable({
  providedIn: 'root'
})
export class TransferService {
  private apiUrl = 'http://localhost:8080/transfer';

  constructor(private http: HttpClient) {}

  createTransfer(request: TransferRequest): Observable<WorldState> {
    return this.http.post<WorldState>(`${this.apiUrl}/create`, request);
  }

  approveTransfer(request: TransferRequest): Observable<WorldState> {
    return this.http.post<WorldState>(`${this.apiUrl}/approve`, request);
  }

  getTransferStatus(transactionId: string): Observable<WorldState> {
    return this.http.get<WorldState>(`${this.apiUrl}/status/${transactionId}`);
  }
}
