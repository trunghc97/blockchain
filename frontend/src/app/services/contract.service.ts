import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { Contract } from '../models/contract.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ContractService {
  private apiUrl = `${environment.apiUrl}/api/v1/contracts`;

  constructor(
    private http: HttpClient,
    private router: Router
  ) {}

  private getHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {})
    });
  }

  private handleError(error: HttpErrorResponse) {
    if (error.status === 401 || error.status === 403) {
      localStorage.removeItem('token');
      this.router.navigate(['/login']);
    }
    return throwError(error);
  }

  getContracts(type?: string, status?: string): Observable<Contract[]> {
    const params: any = {};
    if (type) params.type = type;
    if (status) params.status = status;

    return this.http.get<Contract[]>(this.apiUrl, {
      headers: this.getHeaders(),
      params
    }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  // Legacy method for backward compatibility
  getAllContracts(): Observable<Contract[]> {
    return this.getContracts();
  }

  getLedgerData(contractId: string): Observable<any> {
    return this.http.get(`${this.apiUrl}/${contractId}/ledger`, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  // Transform new API format to old format for frontend compatibility
  transformLedgerData(apiResponse: any): any {
    const events = apiResponse.events || [];

    // Extract unique blocks from events
    const blockMap = new Map();
    events.forEach((event: any) => {
      if (event.blockNumber && !blockMap.has(event.blockNumber)) {
        blockMap.set(event.blockNumber, {
          blockNumber: event.blockNumber,
          hash: event.blockHash,
          merkleRoot: event.merkleRoot,
          timestamp: event.timestamp, // Use event timestamp as block timestamp
          contractEvents: [{
            eventId: event.eventId,
            type: event.type,
            actorId: event.actorId,
            contractId: event.contractId
          }]
        });
      } else if (event.blockNumber) {
        // Add event to existing block
        const block = blockMap.get(event.blockNumber);
        if (block && !block.contractEvents.some((e: any) => e.eventId === event.eventId)) {
          block.contractEvents.push({
            eventId: event.eventId,
            type: event.type,
            actorId: event.actorId,
            contractId: event.contractId
          });
        }
      }
    });

    const blocks = Array.from(blockMap.values());

    return {
      transactions: events,
      blocks: blocks,
      contractId: apiResponse.contractId
    };
  }

  createContract(contract: Partial<Contract>): Observable<Contract> {
    const formData = new FormData();
    formData.append('contract', JSON.stringify(contract));

    return this.http.post<Contract>(this.apiUrl, formData, {
      headers: this.getAuthHeaders()
    }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  createContractWithFile(formData: FormData): Observable<Contract> {
    // Add Authorization header for file upload
    const headers = this.getAuthHeaders();
    return this.http.post<Contract>(this.apiUrl, formData, { headers }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  private getAuthHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    let headers = new HttpHeaders();

    if (token) {
      headers = headers.set('Authorization', `Bearer ${token}`);
    }

    return headers;
  }

  updateContract(id: string, contract: Partial<Contract>): Observable<Contract> {
    return this.http.put<Contract>(`${this.apiUrl}/${id}`, contract, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  approveContract(id: string, approverId?: string): Observable<Contract> {
    const requestBody = approverId ? { supplierId: approverId } : {};
    return this.http.post<Contract>(`${this.apiUrl}/${id}/approve`, requestBody, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  rejectContract(id: string, reason: string): Observable<Contract> {
    return this.http.post<Contract>(`${this.apiUrl}/${id}/reject`, { reason }, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  approveContractByBank(id: string, request: { bankId: string }): Observable<Contract> {
    return this.http.post<Contract>(`${this.apiUrl}/${id}/approve-bank`, request, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  getAllTokens(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tokens/all`, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  getAllBalances(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/balances/all`, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  getBalancesByToken(tokenId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/balances/token/${tokenId}`, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }
}
