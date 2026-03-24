import { Injectable } from '@angular/core';

export interface ToastMessage {
    id: number;
    text: string;
    type: 'success' | 'error' | 'info' | 'warning';
    timestamp: Date;
}

@Injectable({
  providedIn: 'root'
})
export class ToastService {
    toasts: ToastMessage[] = [];
    history: ToastMessage[] = []; 
    private counter = 0;
  
    show(text: string, type: ToastMessage['type'] = 'info', duration = 3000) {
      const id = this.counter++;
      const toast: ToastMessage = { id, text, type, timestamp: new Date() };
      this.toasts.push(toast);
      this.history.push(toast); 
  
      setTimeout(() => this.remove(id), duration);
    }
  
    remove(id: number) {
      this.toasts = this.toasts.filter(t => t.id !== id);
    }

    removeFromHistory(id: number) {
      this.history = this.history.filter(t => t.id !== id);
      this.remove(id);
    }
  
    getToasts(): ToastMessage[] {
      return this.toasts;
    }
  
    getHistory(): ToastMessage[] {
      return this.history;
    }
  
    clearHistory(): void {
      this.history = [];
    }
}
