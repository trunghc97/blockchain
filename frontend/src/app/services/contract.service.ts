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
  private apiUrl = `${environment.apiUrl}/api/contracts`;

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

  getContracts(): Observable<Contract[]> {
    return this.http.get<Contract[]>(this.apiUrl, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  getLedgerData(contractId: string): Observable<any> {
    return this.http.get(`${this.apiUrl}/${contractId}/ledger`, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
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

  approveContract(id: string): Observable<Contract> {
    return this.http.post<Contract>(`${this.apiUrl}/${id}/approve`, {}, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }

  rejectContract(id: string, reason: string): Observable<Contract> {
    return this.http.post<Contract>(`${this.apiUrl}/${id}/reject`, { reason }, { headers: this.getHeaders() }).pipe(
      catchError(this.handleError.bind(this))
    );
  }
}