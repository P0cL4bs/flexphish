import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { ToastService } from 'src/app/services/toast.service';

@Component({
  selector: 'app-toast-container',
  imports: [CommonModule],
  templateUrl: './toast-container.component.html',
  styleUrl: './toast-container.component.css'
})
export class ToastContainerComponent {
  constructor(public toastService: ToastService) { }
}
