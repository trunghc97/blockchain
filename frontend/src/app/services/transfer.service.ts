import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TransferRequest, TransferStatus, ApproveRequest } from '../models/transfer.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class TransferService {
  private apiUrl = `${environment.apiUrl}/transfer`;

  constructor(private http: HttpClient) {}

  createTransfer(request: TransferRequest): Observable<any> {
    return this.http.post(`${this.apiUrl}/create`, request);
  }

  approveTransfer(request: ApproveRequest): Observable<any> {
    return this.http.post(`${this.apiUrl}/approve`, request);
  }

  getTransferList(): Observable<TransferStatus[]> {
    return this.http.get<TransferStatus[]>(`${this.apiUrl}/list`);
  }

  getPendingTransfers(approverId: string): Observable<TransferStatus[]> {
    return this.http.get<TransferStatus[]>(`${this.apiUrl}/list`, {
      params: {
        approverId,
        status: ['PENDING', 'PARTIALLY_APPROVED'].join(',')
      }
    });
  }
}