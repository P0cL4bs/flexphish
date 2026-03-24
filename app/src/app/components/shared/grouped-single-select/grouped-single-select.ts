import { CommonModule } from '@angular/common';
import { Component, ElementRef, EventEmitter, HostListener, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';

export interface GroupedSelectOption {
  label: string;
  value: string | number;
  description?: string;
  searchText?: string;
}

export interface GroupedSelectGroup {
  label: string;
  options: GroupedSelectOption[];
}

@Component({
  selector: 'app-grouped-single-select',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './grouped-single-select.html',
})
export class GroupedSingleSelect {
  @Input() groups: GroupedSelectGroup[] = [];
  @Input() value: string | number | null = null;
  @Input() disabled = false;
  @Input() loading = false;
  @Input() placeholder = 'Select an option';
  @Input() loadingLabel = 'Loading options...';
  @Input() emptyStateLabel = 'No options found';
  @Input() searchPlaceholder = 'Search...';
  @Input() allowClear = true;
  @Input() clearLabel = 'No selection';

  @Output() readonly valueChange = new EventEmitter<string | number | null>();

  search = '';
  isOpen = false;

  constructor(private elementRef: ElementRef<HTMLElement>) {}

  @HostListener('document:click', ['$event'])
  handleDocumentClick(event: MouseEvent): void {
    const target = event.target as Node | null;
    if (!target) {
      return;
    }

    if (!this.elementRef.nativeElement.contains(target)) {
      this.close();
    }
  }

  get selectedLabel(): string {
    for (const group of this.groups) {
      const selected = group.options.find(option => option.value === this.value);
      if (selected) {
        return selected.label;
      }
    }

    return this.placeholder;
  }

  get filteredGroups(): GroupedSelectGroup[] {
    const term = this.search.trim().toLowerCase();
    if (!term) {
      return this.groups;
    }

    return this.groups
      .map(group => ({
        label: group.label,
        options: group.options.filter(option => {
          const haystack = `${option.label} ${option.description || ''} ${option.searchText || ''}`.toLowerCase();
          return haystack.includes(term);
        })
      }))
      .filter(group => group.options.length > 0);
  }

  toggle(): void {
    if (this.disabled) {
      return;
    }

    this.isOpen = !this.isOpen;
    if (!this.isOpen) {
      this.search = '';
    }
  }

  close(): void {
    this.isOpen = false;
    this.search = '';
  }

  select(value: string | number | null): void {
    this.value = value;
    this.valueChange.emit(value);
    this.close();
  }
}
