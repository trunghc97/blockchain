import { Component } from '@angular/core';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  navLinks = [
    { path: '/transfer', label: 'Chuyển tiền' },
    { path: '/approve', label: 'Phê duyệt' },
    { path: '/status', label: 'Trạng thái' }
  ];
}