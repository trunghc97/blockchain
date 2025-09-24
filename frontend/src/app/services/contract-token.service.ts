import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ContractTokenService {
  private apiUrl = `${environment.apiUrl}/api/v1`;

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    return new HttpHeaders({
      'Content-Type': 'application/json'
    });
  }

  // Contract methods
  createContract(contractData: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/contracts`, contractData, { headers: this.getHeaders() });
  }

  approveContract(contractId: string, approver: string): Observable<any> {
    const approvalData = { approver };
    return this.http.post(`${this.apiUrl}/contracts/${contractId}/approve`, approvalData, { headers: this.getHeaders() });
  }

  getContract(contractId: string): Observable<any> {
    return this.http.get(`${this.apiUrl}/contracts/${contractId}`, { headers: this.getHeaders() });
  }

  // Token methods
  getToken(tokenId: string): Observable<any> {
    return this.http.get(`${this.apiUrl}/tokens/${tokenId}`, { headers: this.getHeaders() });
  }

  transferToken(transferData: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/tokens/transfer`, transferData, { headers: this.getHeaders() });
  }

  getTokensIssuedByBank(bankId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tokens/issued/${bankId}`, { headers: this.getHeaders() });
  }

  getSuppliers(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/suppliers`, { headers: this.getHeaders() });
  }

  getAllTokens(): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tokens`, { headers: this.getHeaders() });
  }

  getBalancesByAccount(accountId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tokens/balances/account/${accountId}`, { headers: this.getHeaders() });
  }

  getBalancesByToken(tokenId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.apiUrl}/tokens/balances/token/${tokenId}`, { headers: this.getHeaders() });
  }

  settleToken(settleData: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/tokens/settle`, settleData, { headers: this.getHeaders() });
  }
}
