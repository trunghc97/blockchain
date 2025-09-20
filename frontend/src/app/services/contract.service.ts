import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Contract } from '../models/contract.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class ContractService {
  constructor(private http: HttpClient) {}

  createContract(formData: FormData): Observable<Contract> {
    return this.http.post<Contract>(`${environment.apiUrl}/api/contracts`, formData);
  }

  getContracts(): Observable<Contract[]> {
    return this.http.get<Contract[]>(`${environment.apiUrl}/api/contracts`);
  }

  getContract(id: string): Observable<Contract> {
    return this.http.get<Contract>(`${environment.apiUrl}/api/contracts/${id}`);
  }

  approveContract(id: string): Observable<Contract> {
    return this.http.post<Contract>(`${environment.apiUrl}/api/contracts/${id}/approve`, {});
  }
}