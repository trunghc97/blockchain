import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { UserService } from '../../services/user.service';
import { firstValueFrom } from 'rxjs';

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
    private userService: UserService,
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
        next: async (response) => {
          console.log('Login successful:', response);
          this.loading = false;

          // Redirect based on user role
          try {
            const currentUser = await firstValueFrom(this.userService.getCurrentUser());
            console.log('Current user role:', currentUser?.role);

            if (currentUser?.role === 'BANK') {
              this.router.navigate(['/bank']);
            } else if (currentUser?.role === 'SUPPLIER') {
              this.router.navigate(['/supplier']);
            } else {
              // Default to contracts for ANCHOR or unknown roles
              this.router.navigate(['/contracts']);
            }
          } catch (error) {
            console.error('Error getting current user after login:', error);
            // Fallback to contracts
            this.router.navigate(['/contracts']);
          }
        },
        error: (error) => {
          console.error('Login error:', error);
          this.error = error.error?.message || 'Đăng nhập thất bại';
          this.loading = false;
        }
      });
  }
}
