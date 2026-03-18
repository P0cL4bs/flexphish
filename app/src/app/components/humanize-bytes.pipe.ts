import { Pipe, type PipeTransform } from '@angular/core';

@Pipe({
  name: 'humanizeBytes'
})
export class HumanizeBytesPipe implements PipeTransform {

  transform(value: number): string {
    if (value === 0 || Number.isNaN(value)) return '0 B';
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(value) / Math.log(1024));
    // biome-ignore lint/style/useExponentiationOperator: <explanation>
    const size = value / Math.pow(1024, i);
    return `${size.toFixed(1)} ${sizes[i]}`;
  }
}
