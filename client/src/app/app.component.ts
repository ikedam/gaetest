import { Component } from '@angular/core';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'gaetest';
  navbarCollapsed = true;
  entityCollapsed = true;

  collapseAll() {
    this.navbarCollapsed = true;
    this.entityCollapsed = true;
  }
}
