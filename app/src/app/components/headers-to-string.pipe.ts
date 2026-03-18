import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
    name: 'headersToString'
})
export class HeadersToStringPipe implements PipeTransform {
    transform(headers: Record<string, any> | null | undefined): string {
        if (!headers) return '';

        // headers normalmente têm formato: { "Name": ["v1","v2"], ... }
        try {
            return Object.entries(headers)
                .map(([k, v]) => {
                    // se for array, juntar por ', '; se for string, usar direto
                    const values = Array.isArray(v) ? v.join(', ') : String(v);
                    return `${k}: ${values}`;
                })
                .join('\n');
        } catch (e) {
            // fallback: stringify
            return JSON.stringify(headers, null, 2);
        }
    }
}