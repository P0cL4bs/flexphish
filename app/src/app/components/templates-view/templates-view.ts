import { Component, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TemplateMetadata } from 'src/app/models/template.model';
import { ApiService } from 'src/app/services/api.service';
import { yaml } from '@codemirror/lang-yaml';
import { html } from '@codemirror/lang-html';
import { CodeEditor } from '@acrodata/code-editor';
import { ActivatedRoute, NavigationEnd, Router, RouterOutlet } from "@angular/router";
import { debounceTime, filter, Subject } from 'rxjs';
import { TemplateCreateView } from '../template-create-view/template-create-view';
import { LucideAngularModule } from "lucide-angular";
import { ToastService } from 'src/app/services/toast.service';

@Component({
  selector: 'app-templates-view',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterOutlet,
    TemplateCreateView,
    LucideAngularModule
  ],
  templateUrl: './templates-view.html',
  styleUrl: './templates-view.css'
})
export class TemplatesView implements OnInit {

  activeTemplate: string | null = null;
  templates: TemplateMetadata[] = [];
  selectedTemplate: TemplateMetadata | null = null;

  search = ''
  selectedCategory: string | null = null
  selectedTag: string | null = null
  sortBy: 'name' | 'category' = 'name'

  filteredTemplates: TemplateMetadata[] = []

  @ViewChild(TemplateCreateView)
  createView!: TemplateCreateView;

  constructor(private api: ApiService, private router: Router,
    private route: ActivatedRoute, private toastr: ToastService) { }

  ngOnInit(): void {
    this.loadTemplates();
    this.router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .subscribe(() => {

        const child = this.route.firstChild;

        if (child) {
          const filename = child.snapshot.paramMap.get('filename');

          if (filename) {
            this.activeTemplate = filename + '.yaml';
          }
        }

      });

    this.route.firstChild?.paramMap.subscribe(params => { const filename = params.get('filename'); if (filename) { this.activeTemplate = filename + '.yaml'; } });

    window.addEventListener('templates:reload', () => {
      this.loadTemplates();
    });

  }


  applyFilter() {

    let list = [...this.templates]

    const search = this.search.toLowerCase()

    if (search) {
      list = list.filter(t =>
        t.filename.toLowerCase().includes(search) ||
        t.category?.toLowerCase().includes(search) ||
        t.tags?.some(tag => tag.toLowerCase().includes(search))
      )
    }

    if (this.selectedCategory) {
      list = list.filter(t => t.category === this.selectedCategory)
    }

    if (this.selectedTag) {
      list = list.filter(t => t.tags?.includes(this.selectedTag!))
    }

    if (this.sortBy === 'name') {
      list.sort((a, b) => a.filename.localeCompare(b.filename))
    }

    if (this.sortBy === 'category') {
      list.sort((a, b) => (a.category || '').localeCompare(b.category || ''))
    }

    this.filteredTemplates = list

  }

  get categories(): string[] {
    return [...new Set(
      this.templates
        .map(t => t.category)
        .filter(Boolean)
    )] as string[]
  }

  get tags(): string[] {

    const tags = this.templates.flatMap(t => t.tags || [])

    return [...new Set(tags)]

  }

  openCreateTemplate() {

    this.createView.openModal();

  }

  selectTemplate(template: TemplateMetadata): void {

    this.router.navigate([
      '/templates',
      template.filename.replace('.yaml', '')
    ]);

  }

  loadTemplates(): void {
    this.api.getTemplatesList().subscribe({
      next: (data) => {
        this.templates = data;
        this.filteredTemplates = data
      },
      error: (err) => {
        console.error('Failed loading templates', err);
        this.toastr.show(err, "error")
      }
    });
  }

  filterTemplates() {

    const term = this.search.toLowerCase().trim()

    if (!term) {
      this.filteredTemplates = this.templates
      return
    }

    this.filteredTemplates = this.templates.filter(t =>
      t.filename.toLowerCase().includes(term) ||
      t.category?.toLowerCase().includes(term) ||
      t.tags?.some(tag => tag.toLowerCase().includes(term))
    )

  }
}