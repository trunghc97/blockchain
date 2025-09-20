import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { LoginComponent } from './components/login/login.component';
import { ContractFormComponent } from './components/contract-form/contract-form.component';
import { ContractApprovalComponent } from './components/contract-approval/contract-approval.component';
import { ContractStatusComponent } from './components/contract-status/contract-status.component';
import { LedgerViewerComponent } from './components/ledger-viewer/ledger-viewer.component';
import { AuthGuard } from './guards/auth.guard';

const routes: Routes = [
  { path: 'login', component: LoginComponent },
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
    path: 'contracts/status', 
    component: ContractStatusComponent,
    canActivate: [AuthGuard]
  },
  { 
    path: 'ledger', 
    component: LedgerViewerComponent,
    canActivate: [AuthGuard]
  },
  { path: '', redirectTo: '/contracts/new', pathMatch: 'full' },
  { path: '**', redirectTo: '/contracts/new' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }