import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ContractTokenService {
  private contractsApiUrl = `${environment.apiUrl}/api/v1/contracts`;
  private tokensApiUrl = `${environment.apiUrl}/api/v1/tokens`;
  private suppliersApiUrl = `${environment.apiUrl}/api/v1/suppliers`;

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    return new HttpHeaders({
      'Content-Type': 'application/json'
    });
  }

  // Contract methods
  createContract(contractData: any): Observable<any> {
    return this.http.post(this.contractsApiUrl, contractData, { headers: this.getHeaders() });
  }

  approveContract(contractId: string, approver: string): Observable<any> {
    const approvalData = { supplierId: approver };
    return this.http.post(`${this.contractsApiUrl}/${contractId}/approve`, approvalData, { headers: this.getHeaders() });
  }

  getContract(contractId: string): Observable<any> {
    return this.http.get(`${this.contractsApiUrl}/${contractId}`, { headers: this.getHeaders() });
  }

  // Token methods
  getToken(tokenId: string): Observable<any> {
    return this.http.get(`${this.tokensApiUrl}/${tokenId}`, { headers: this.getHeaders() });
  }

  transferToken(transferData: any): Observable<any> {
    return this.http.post(`${this.tokensApiUrl}/transfer`, transferData, { headers: this.getHeaders() });
  }

  getTokensIssuedByBank(bankId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.tokensApiUrl}/issued/${bankId}`, { headers: this.getHeaders() });
  }

  getSuppliers(): Observable<any[]> {
    return this.http.get<any[]>(this.suppliersApiUrl, { headers: this.getHeaders() });
  }

  getAllTokens(): Observable<any[]> {
    return this.http.get<any[]>(this.tokensApiUrl, { headers: this.getHeaders() });
  }

  getContracts(): Observable<any[]> {
    return this.http.get<any[]>(this.contractsApiUrl, { headers: this.getHeaders() });
  }

  getBalancesByAccount(accountId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.tokensApiUrl}/balances/account/${accountId}`, { headers: this.getHeaders() });
  }

  getBalancesByToken(tokenId: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.tokensApiUrl}/balances/token/${tokenId}`, { headers: this.getHeaders() });
  }

  settleToken(settleData: any): Observable<any> {
    return this.http.post(`${this.tokensApiUrl}/settle`, settleData, { headers: this.getHeaders() });
  }
}
