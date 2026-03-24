import { CommonModule } from '@angular/common';
import { Component, ElementRef, HostListener, type OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { ApiService } from 'src/app/services/api.service';
import { SidebarService } from 'src/app/services/sidebar.service';
import { ToastMessage, ToastService } from 'src/app/services/toast.service';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-navbar',
  imports: [CommonModule, LucideAngularModule],
  templateUrl: './navbar.component.html',
  styleUrl: './navbar.component.css'
})
export class NavbarComponent implements OnInit {
  isVisibleCommandBar: boolean = false
  isNotificationsOpen = false;
  readonly maxNotifications = 8;

  env: any = environment;
  constructor(
    public api: ApiService,
    public sidebarService: SidebarService,
    public toastService: ToastService,
    private elementRef: ElementRef<HTMLElement>,
    private router: Router
  ) {
  }

  toggleCommandBar() {
  }

  toggleSidebar() {
    this.sidebarService.toggle();
  }


  async ngOnInit() {
  }

  get notifications(): ToastMessage[] {
    return [...this.toastService.getHistory()]
      .sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
      .slice(0, this.maxNotifications);
  }

  get notificationCount(): number {
    return this.toastService.getHistory().length;
  }

  clearNotifications() {
    this.toastService.clearHistory();
    this.isNotificationsOpen = true;
  }

  removeNotification(id: number) {
    this.toastService.removeFromHistory(id);
    this.isNotificationsOpen = true;
  }

  toggleNotifications(event: MouseEvent) {
    event.stopPropagation();
    this.isNotificationsOpen = !this.isNotificationsOpen;
  }

  keepNotificationsOpen(event: MouseEvent) {
    event.stopPropagation();
    this.isNotificationsOpen = true;
  }

  @HostListener('document:click', ['$event'])
  onDocumentClick(event: MouseEvent) {
    const target = event.target as Node | null;
    if (!target) return;

    const clickedInside = this.elementRef.nativeElement
      .querySelector('[data-notification-dropdown]')
      ?.contains(target);

    if (!clickedInside) {
      this.isNotificationsOpen = false;
    }
  }

  getNotificationBadgeClass(type: ToastMessage['type']): string {
    const badgeByType = {
      success: 'badge-success',
      error: 'badge-error',
      warning: 'badge-warning',
      info: 'badge-info'
    };
    return badgeByType[type];
  }

  formatNotificationTime(timestamp: Date): string {
    const now = Date.now();
    const diffMs = Math.max(0, now - new Date(timestamp).getTime());
    const diffMinutes = Math.floor(diffMs / 60000);
    if (diffMinutes < 1) return 'now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;

    const diffHours = Math.floor(diffMinutes / 60);
    if (diffHours < 24) return `${diffHours}h ago`;

    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
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
