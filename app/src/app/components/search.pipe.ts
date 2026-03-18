import { Pipe, type PipeTransform } from '@angular/core';

@Pipe({
  standalone: true,
  name: 'search'
})
export class SearchPipe implements PipeTransform {
  transform(items: any[], query: string): any[] {
    if (!items || !query) return items;
    query = query.toLowerCase();
    return items.filter(item =>
      JSON.stringify(item).toLowerCase().includes(query)
    );
  }
}
