import { CommonModule, DOCUMENT } from '@angular/common'
import { Component, type ElementRef, Inject, OnInit, ViewChild } from '@angular/core'
import type { Title } from '@angular/platform-browser'
// biome-ignore lint/style/useImportType: <explanation>
import { ActivatedRoute, Route, Router, RouterOutlet } from '@angular/router'
import { select } from '@ngneat/elf'
import { LucideAngularModule } from 'lucide-angular'
import { Subscription } from 'rxjs'
import { NavbarComponent } from './components/core/navbar/navbar.component'
import { SidebarComponent } from './components/core/sidebar/sidebar.component'
import { ToastContainerComponent } from './components/core/toast-container/toast-container.component'
// biome-ignore lint/style/useImportType: <explanation>
import { ApiService } from './services/api.service'
// biome-ignore lint/style/useImportType: <explanation>
import { ToastService } from './services/toast.service'
import { THEMES, themeStore } from './stores/theme.store'
import { HttpErrorResponse } from '@angular/common/http'

@Component({
  selector: 'app-root',
  standalone: true,
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
  imports: [SidebarComponent, NavbarComponent, ToastContainerComponent, RouterOutlet, CommonModule],
})
export class AppComponent implements OnInit {
  theme = 'dark'
  themeSubscription: Subscription = new Subscription()
  apiDisconnected = false;

  constructor(public api: ApiService, private router: Router, public route: ActivatedRoute, private toast: ToastService) {
    console.log("AppComponent()");
    this.api.onLoggedIn.subscribe(() => {
      console.log("logged in");
      this.apiDisconnected = false;
    });

    this.api.onLoggedOut.subscribe(error => {
      console.log("logged out");
      this.api.logout();
      this.router.navigateByUrl("/login");
    });
  }


  ngOnInit() {
    console.log("loading")
    this.themeSubscription = themeStore.pipe(select((state) => state.theme)).subscribe((s) => this.setTheme(s?.isDark))
  }

  ngDestory() {
    this.themeSubscription.unsubscribe()
  }

  private setTheme(isDark?: boolean) {
    document.body.setAttribute('data-theme', isDark ? THEMES.DARK : THEMES.LIGHT)
  }
}
