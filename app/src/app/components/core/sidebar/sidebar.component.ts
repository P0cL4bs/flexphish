import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterModule } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { SidebarService } from 'src/app/services/sidebar.service';

@Component({
  selector: 'app-sidebar',
  imports: [LucideAngularModule, RouterModule, CommonModule],
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.css'
})
export class SidebarComponent implements OnInit {
  isVisibleSidebar: boolean = true

  constructor(public sidebarService: SidebarService) {
  }

  toggleSidebar() {
    this.sidebarService.toggle();
  }

  async ngOnInit() {
    this.sidebarService.visible$.subscribe(v => {
      this.isVisibleSidebar = v;
    });
  }
}
