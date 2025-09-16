import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { TransferFormComponent } from './components/transfer-form/transfer-form.component';
import { ApproverComponent } from './components/approver/approver.component';
import { StatusListComponent } from './components/status-list/status-list.component';

const routes: Routes = [
  { path: 'transfer', component: TransferFormComponent },
  { path: 'approve', component: ApproverComponent },
  { path: 'status', component: StatusListComponent },
  { path: '', redirectTo: '/transfer', pathMatch: 'full' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
