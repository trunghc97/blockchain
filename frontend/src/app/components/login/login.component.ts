import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  loginForm: FormGroup;
  error: string = '';
  loading: boolean = false;

  constructor(
    private formBuilder: FormBuilder,
    private authService: AuthService,
    private router: Router
  ) {
    this.loginForm = this.formBuilder.group({
      username: ['', Validators.required],
      password: ['', Validators.required]
    });
  }

  onSubmit(): void {
    if (this.loginForm.invalid) {
      return;
    }

    this.loading = true;
    this.error = '';

    this.authService.login(this.loginForm.value)
      .subscribe({
        next: (response) => {
          console.log('Login successful:', response);
          this.loading = false;
          // Redirect trực tiếp đến trang contracts
          this.router.navigate(['/contracts']);
        },
        error: (error) => {
          console.error('Login error:', error);
          this.error = error.error?.message || 'Đăng nhập thất bại';
          this.loading = false;
        }
      });
  }
}