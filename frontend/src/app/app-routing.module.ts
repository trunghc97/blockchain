import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AuthGuard } from './guards/auth.guard';

// Import các component trực tiếp
import { LoginComponent } from './components/login/login.component';
import { ContractStatusComponent } from './components/contract-status/contract-status.component';
import { ContractFormComponent } from './components/contract-form/contract-form.component';
import { ContractApprovalComponent } from './components/contract-approval/contract-approval.component';
import { LedgerViewerComponent } from './components/ledger-viewer/ledger-viewer.component';

// Contract-Token Management Components
import { AnchorComponent } from './components/contract-token-management/anchor/anchor.component';
import { BankComponent } from './components/contract-token-management/bank/bank.component';
import { SupplierComponent } from './components/contract-token-management/supplier/supplier.component';

const routes: Routes = [
  {
    path: '',
    redirectTo: '/contracts',
    pathMatch: 'full'
  },
  {
    path: 'login',
    component: LoginComponent
  },
  // Contract-Token Management Routes
  {
    path: 'bank',
    component: BankComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'supplier',
    component: SupplierComponent,
    canActivate: [AuthGuard]
  },
  // Create contract route
  {
    path: 'contracts/new',
    component: ContractFormComponent,
    canActivate: [AuthGuard]
  },
  // View contracts status
  {
    path: 'contracts',
    component: ContractStatusComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'contracts/approve',
    component: ContractApprovalComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'ledger',
    component: LedgerViewerComponent,
    canActivate: [AuthGuard]
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }