import { CommonModule } from '@angular/common';
import { Component, ViewChild, type OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { firstValueFrom } from 'rxjs/internal/firstValueFrom';
import { ApiService } from 'src/app/services/api.service';
import { SidebarService } from 'src/app/services/sidebar.service';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-navbar',
  imports: [CommonModule, LucideAngularModule],
  templateUrl: './navbar.component.html',
  styleUrl: './navbar.component.css'
})
export class NavbarComponent implements OnInit {
  isVisibleCommandBar: boolean = false

  env: any = environment;
  constructor(public api: ApiService, public sidebarService: SidebarService, private router: Router) {
  }

  toggleCommandBar() {
  }

  toggleSidebar() {
    this.sidebarService.toggle();
  }


  async ngOnInit() {
  }

  logout() {
    const confirmed = window.confirm(
      "Are you sure you want to log out?."
    );
    if (confirmed) {
      this.api.onLoggedOut.emit();
      this.router.navigateByUrl("/login");
    }
  }
}
