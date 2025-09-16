import { Component } from '@angular/core';

@Component({
  selector: 'app-root',
  template: `
    <div class="container">
      <h1>Blockchain Transfer System</h1>
      <div class="row">
        <div class="col">
          <app-transfer-form></app-transfer-form>
        </div>
        <div class="col">
          <app-approver></app-approver>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .container {
      padding: 20px;
      max-width: 1200px;
      margin: 0 auto;
    }
    h1 {
      text-align: center;
      margin-bottom: 30px;
    }
    .row {
      display: flex;
      gap: 20px;
    }
    .col {
      flex: 1;
    }
    @media (max-width: 768px) {
      .row {
        flex-direction: column;
      }
    }
  `]
})
export class AppComponent {
  title = 'blockchain-frontend';
}
