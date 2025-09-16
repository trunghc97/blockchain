import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpClientModule } from '@angular/common/http';
import { ReactiveFormsModule } from '@angular/forms';

import { AppComponent } from './app.component';
import { TransferFormComponent } from './components/transfer-form/transfer-form.component';
import { ApproverComponent } from './components/approver/approver.component';

@NgModule({
  declarations: [
    AppComponent,
    TransferFormComponent,
    ApproverComponent
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    ReactiveFormsModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
