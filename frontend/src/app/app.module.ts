import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { CommonModule, DatePipe } from '@angular/common';

// Routing
import { RouterModule } from '@angular/router';

// Material Modules
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatTableModule } from '@angular/material/table';
import { MatSortModule } from '@angular/material/sort';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTabsModule } from '@angular/material/tabs';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatDialogModule } from '@angular/material/dialog';
import { MatOptionModule } from '@angular/material/core';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatGridListModule } from '@angular/material/grid-list';

// Routing
import { AppRoutingModule } from './app-routing.module';

// Interceptors
import { ErrorInterceptor } from './interceptors/error.interceptor';

// Components
import { AppComponent } from './app.component';
import { ContractStatusComponent } from './components/contract-status/contract-status.component';
import { TokenDetailsDialogComponent } from './components/contract-status/token-details-dialog.component';
import { LedgerViewerComponent } from './components/ledger-viewer/ledger-viewer.component';
import { ContractFormComponent } from './components/contract-form/contract-form.component';
import { ContractApprovalComponent } from './components/contract-approval/contract-approval.component';
import { LoginComponent } from './components/login/login.component';
import { NavbarComponent } from './components/navbar/navbar.component';

// Contract-Token Management Components
import { AnchorComponent } from './components/contract-token-management/anchor/anchor.component';
import { BankComponent } from './components/contract-token-management/bank/bank.component';
import { SupplierComponent } from './components/contract-token-management/supplier/supplier.component';

@NgModule({
  declarations: [
    AppComponent,
    ContractStatusComponent,
    TokenDetailsDialogComponent,
    LedgerViewerComponent,
    ContractFormComponent,
    ContractApprovalComponent,
    LoginComponent,
    NavbarComponent,
    // Contract-Token Management Components
    AnchorComponent,
    BankComponent,
    SupplierComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    CommonModule,
    HttpClientModule,
    AppRoutingModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    // Material Modules
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatSelectModule,
    MatTableModule,
    MatSortModule,
    MatPaginatorModule,
    MatChipsModule,
    MatProgressSpinnerModule,
    MatTabsModule,
    MatTooltipModule,
    MatSnackBarModule,
    MatToolbarModule,
    MatExpansionModule,
    MatButtonToggleModule,
    MatAutocompleteModule,
    MatDialogModule,
    MatOptionModule,
    MatListModule,
    MatMenuModule,
    MatSidenavModule,
    MatGridListModule
  ],
  providers: [
    DatePipe, // Add DatePipe provider
    {
      provide: HTTP_INTERCEPTORS,
      useClass: ErrorInterceptor,
      multi: true
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
