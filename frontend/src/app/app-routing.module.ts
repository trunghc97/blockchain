import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AuthGuard } from './guards/auth.guard';

// Import các component trực tiếp
import { LoginComponent } from './components/login/login.component';
import { ContractStatusComponent } from './components/contract-status/contract-status.component';
import { ContractFormComponent } from './components/contract-form/contract-form.component';
import { ContractApprovalComponent } from './components/contract-approval/contract-approval.component';
import { LedgerViewerComponent } from './components/ledger-viewer/ledger-viewer.component';

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
  {
    path: 'contracts',
    component: ContractStatusComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'contracts/new',
    component: ContractFormComponent,
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