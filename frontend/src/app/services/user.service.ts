import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { tap, catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';
import { User } from '../models/user.model';

@Injectable({
  providedIn: 'root'
})
export class UserService {
  private apiUrl = `${environment.apiUrl}/api/users`;
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  currentUser$ = this.currentUserSubject.asObservable();

  constructor(
    private http: HttpClient,
    private router: Router
  ) {
    this.loadCurrentUser();
  }

  private loadCurrentUser() {
    const token = localStorage.getItem('token');
    if (token) {
      this.getCurrentUser().subscribe({
        next: (user) => this.currentUserSubject.next(user),
        error: () => {
          // Nếu không lấy được user, xóa token
          localStorage.removeItem('token');
          this.currentUserSubject.next(null);
        }
      });
    }
  }

  private getHeaders(): HttpHeaders {
    const token = localStorage.getItem('token');
    return new HttpHeaders({
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {})
    });
  }

  // Phương thức để cập nhật current user từ bên ngoài
  updateCurrentUser(user: User) {
    this.currentUserSubject.next(user);
  }

  getCurrentUser(): Observable<User> {
    return this.http.get<User>(`${this.apiUrl}/current`, { headers: this.getHeaders() }).pipe(
      tap(user => this.currentUserSubject.next(user)),
      catchError((error) => {
        if (error.status === 401 || error.status === 403) {
          this.logout();
        }
        return throwError(error);
      })
    );
  }

  getSuppliers(): Observable<User[]> {
    return this.http.get<User[]>(`${this.apiUrl}/suppliers`, { headers: this.getHeaders() }).pipe(
      catchError((error) => {
        if (error.status === 401 || error.status === 403) {
          this.logout();
        }
        return throwError(error);
      })
    );
  }

  logout() {
    // Xóa token từ localStorage
    localStorage.removeItem('token');
    // Reset current user
    this.currentUserSubject.next(null);
    // Chuyển hướng về trang login
    this.router.navigate(['/login']);
  }
}