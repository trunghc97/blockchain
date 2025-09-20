import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent, HttpErrorResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
  constructor(
    private router: Router,
    private snackBar: MatSnackBar
  ) {}

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(request).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 401 || error.status === 403) {
          // Xóa token khỏi localStorage
          localStorage.removeItem('token');
          
          // Hiển thị thông báo
          this.snackBar.open('Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.', 'Đóng', {
            duration: 3000,
            horizontalPosition: 'center',
            verticalPosition: 'bottom'
          });

          // Chuyển hướng về trang đăng nhập
          this.router.navigate(['/login']);
        }

        return throwError(error);
      })
    );
  }
}
