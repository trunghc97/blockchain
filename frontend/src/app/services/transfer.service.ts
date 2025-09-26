import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { TransferRequest, TransferStatus, ApproveRequest } from '../models/transfer.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class TransferService {
  private apiUrl = `${environment.apiUrl}/api/v1/tokens`;

  constructor(private http: HttpClient) {}

  createTransfer(request: TransferRequest): Observable<any> {
    // Map TransferRequest to token transfer format
    const transferData = {
      tokenId: request.tokenId,
      from: request.from,
      to: request.to,
      amount: request.amount
    };
    return this.http.post(`${this.apiUrl}/transfer`, transferData);
  }

  approveTransfer(request: ApproveRequest): Observable<any> {
    // Transfer approval is handled through token settlement
    const settleData = {
      transferId: request.transferId,
      approverId: request.approverId
    };
    return this.http.post(`${this.apiUrl}/settle`, settleData);
  }

  getTransferList(): Observable<TransferStatus[]> {
    // Get all tokens as transfers are represented as token movements
    return this.http.get<TransferStatus[]>(this.apiUrl);
  }

  sortTransfersByDate(transfers: TransferStatus[]): TransferStatus[] {
    return transfers.sort((a, b) => {
      const dateA = new Date(a.lastUpdated).getTime();
      const dateB = new Date(b.lastUpdated).getTime();
      return dateB - dateA; // Sắp xếp giảm dần (mới nhất lên đầu)
    });
  }

  getPendingTransfers(approverId: string): Observable<TransferStatus[]> {
    // For now, return all tokens - in real implementation this would filter by status
    return this.http.get<TransferStatus[]>(this.apiUrl);
  }
}
