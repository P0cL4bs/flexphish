import {EventEmitter, Injectable } from '@angular/core';

export interface ColumnSortedEvent {
    field: string;
    direction: string;
    type: string;
}

@Injectable({
    providedIn: 'root'
})
export class SortService {

    public onSort: EventEmitter<ColumnSortedEvent> = new EventEmitter();

    emitSort(event: ColumnSortedEvent) {
        this.onSort.emit(event);
    }

    sort<T extends Record<string, any>>(array: T[], how: ColumnSortedEvent): void {
        const { field, direction, type } = how;
        const less = direction === 'desc' ? -1 : 1;
        const greater = -less;
      
        array.sort((a, b) => {
          const valA = a[field];
          const valB = b[field];
      
          if (type === 'ip') {
            const ipA = valA.split('.').map((octet: string) => Number.parseInt(octet, 10));
            const ipB = valB.split('.').map((octet: string) => Number.parseInt(octet, 10));
      
            for (let i = 0; i < ipA.length; i++) {
              if (ipA[i] < ipB[i]) return less;
              if (ipA[i] > ipB[i]) return greater;
            }
            return 0;
          }
      
          if (valA < valB) return less;
          if (valA > valB) return greater;
          return 0;
        });
      }
}

