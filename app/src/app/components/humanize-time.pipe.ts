import { Pipe, PipeTransform } from '@angular/core';

@Pipe({ name: 'humanizeTime' })
export class HumanTimePipe implements PipeTransform {
    transform(seconds: number): string {
        if (!seconds || seconds < 1) return '0s';
        const m = Math.floor(seconds / 60);
        const s = seconds % 60;
        return m > 0 ? `${m}m ${s}s` : `${s}s`;
    }
}