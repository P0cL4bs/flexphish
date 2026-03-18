import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { ApiService } from 'src/app/services/api.service';
import { Template, TemplateHtmlFile, TemplateHtmlFileUpdateRequest, TemplateMetadata, TemplateStaticFile, TemplateStaticFileRequest, TemplateUpdateRequest } from 'src/app/models/template.model';

import { yaml } from '@codemirror/lang-yaml';
import { html } from '@codemirror/lang-html';
import { css } from '@codemirror/lang-css';
import { javascript } from '@codemirror/lang-javascript';
import { CodeEditor } from '@acrodata/code-editor';
import { LucideAngularModule } from "lucide-angular";
import { ToastService } from 'src/app/services/toast.service';

@Component({
  selector: 'app-template-view-by-id',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    CodeEditor,
    LucideAngularModule
  ],
  templateUrl: './template-view-by-id.html'
})
export class TemplateViewByID implements OnInit {

  language_yaml = yaml();
  language_html = html();

  templateMetadata: TemplateMetadata | null = null;
  template: Template | null = null;

  editingFile: any = null;
  editingFileContent = '';
  loading = true;

  showDeleteModal = false;
  htmlFiles: TemplateHtmlFile[] = [];
  originalFileContent = '';
  staticFiles: TemplateStaticFile[] = [];
  editingStaticFile: any = null;
  editingStaticFileContent = '';
  originalStaticFileContent = '';

  constructor(
    private api: ApiService,
    private route: ActivatedRoute, private router: Router, private toastr: ToastService
  ) { }

  ngOnInit(): void {

    this.route.paramMap.subscribe(params => {

      const filename = params.get('filename') + '.yaml';
      if (!filename) return;

      this.loadTemplate(filename);
      this.loadHtmlFiles(filename);
      this.loadStaticFiles(filename);
    });

  }

  loadTemplate(filename: string) {

    this.api.getTemplateMetadataById(filename).subscribe({
      next: (data) => {
        this.templateMetadata = data;
      },
      error: (err) => {
        console.error('Error loading template', err);
        this.toastr.show(err, "error")
      }
    });

    this.api.getTemplateById(filename).subscribe({
      next: (data) => {
        this.template = data;
        this.loading = false;
      },
      error: (err) => {
        console.error('Error loading template', err);
        this.toastr.show(err, "error")
      }
    });

  }

  loadHtmlFiles(filename: string) {

    this.api.getTemplateHtmlFiles(filename).subscribe({

      next: (files) => {
        this.htmlFiles = files;
      },

      error: (err) => {
        console.error('Failed loading html files', err);
        this.toastr.show(err, "error")
      }

    });

  }

  loadStaticFiles(filename: string) {

    this.api.getTemplateStaticFiles(filename).subscribe({

      next: (files) => {
        this.staticFiles = files;
      },

      error: (err) => {
        console.error('Failed loading static files', err);
        this.toastr.show(err, "error");
      }

    });

  }

  saveTemplate(): void {

    if (!this.templateMetadata) return;

    const req: TemplateUpdateRequest = {
      filename: this.templateMetadata.filename,
      content: this.templateMetadata.content
    };

    this.api.updateTemplate(req).subscribe({
      next: () => {
        console.log("Template updated successfully");
        this.toastr.show("Template updated successfully.", "success");
      },
      error: (err) => {
        const message = err?.error?.error || "Failed to update template";
        this.toastr.show(message, "error");
      }
    });

  }

  openDeleteModal() {
    this.showDeleteModal = true;
  }

  closeDeleteModal() {
    this.showDeleteModal = false;
  }


  confirmDelete() {

    if (!this.templateMetadata) return;

    this.api.deleteTemplate({
      filename: this.templateMetadata.filename
    }).subscribe({
      next: () => {
        this.showDeleteModal = false;
        this.toastr.show("The Template has been deleted successfully.", "success");
        window.dispatchEvent(new Event('templates:reload'));
      },
      error: (err) => {
        const message = err?.error?.error || "Failed to delete";
        this.toastr.show(message, "error");
      }
    });

  }

  onFileUpload(event: any): void {

    const files: FileList = event.target.files;

    if (!files || !this.templateMetadata) return;

    Array.from(files).forEach(file => {

      const reader = new FileReader();

      reader.onload = () => {

        const payload = {
          t_filename: this.templateMetadata!.filename,
          filename: file.name,
          content: reader.result as string
        };

        this.api.uploadTemplateHtmlFile(payload, this.templateMetadata!.filename)
          .subscribe({
            next: () => {

              console.log('File uploaded:', file.name);

              this.loadHtmlFiles(this.templateMetadata!.filename);
              this.toastr.show("The file has been updated successfully.", "success");
            },
            error: (err) => {
              const message = err?.error?.error || "Failed to upload file";
              this.toastr.show(message, "error");
            }
          });

      };

      reader.readAsText(file);

    });

  }

  onEditFile(file: any): void {

    this.editingFile = file;
    this.editingFileContent = file.content;
    this.originalFileContent = file.content;

  }

  hasChanges(): boolean {
    return this.editingFileContent !== this.originalFileContent;
  }

  saveEditedFile(): void {

    if (!this.editingFile || !this.templateMetadata) return;

    const payload: TemplateHtmlFileUpdateRequest = {
      t_filename: this.templateMetadata.filename,
      filename: this.editingFile.filename,
      content: this.editingFileContent
    };

    this.api.updateTemplateHtmlFile(payload, this.templateMetadata.filename).subscribe({
      next: () => {

        this.editingFile.content = this.editingFileContent;

        this.editingFile = null;
        this.editingFileContent = '';

        console.log('HTML file updated');
        this.toastr.show("The HTML file has been updated.", "success");

      },
      error: (err) => {
        const message = err?.error?.error || "Failed to upload file";
        this.toastr.show(message, "error");
      }
    });

  }
  onDeleteFile(file: TemplateHtmlFile): void {

    if (!this.templateMetadata) return;

    const payload = {
      t_filename: this.templateMetadata.filename,
      filename: file.filename
    };

    this.api.deleteTemplateHtmlFile(payload, this.templateMetadata.filename)
      .subscribe({
        next: () => {

          console.log('File deleted:', file.filename);

          this.loadHtmlFiles(this.templateMetadata!.filename);
          this.toastr.show("The file has been deleted successfully.", "success");
        },
        error: (err) => {
          const message = err?.error?.error || "Failed to upload file";
          this.toastr.show(message, "error");
        }
      });

  }

  goBack(): void {
    this.router.navigate(['/templates']);
  }
  onStaticFileUpload(event: any): void {

    const files: FileList = event.target.files;

    if (!files || !this.templateMetadata) return;

    Array.from(files).forEach(file => {

      const reader = new FileReader();

      reader.onload = () => {

        const base64 = (reader.result as string).split(',')[1];

        const payload: TemplateStaticFileRequest = {
          t_filename: this.templateMetadata!.filename,
          filename: file.name,
          content: base64
        };

        this.api.createTemplateStaticFile(payload)
          .subscribe({
            next: () => {

              console.log('Static file uploaded:', file.name);

              this.loadStaticFiles(this.templateMetadata!.filename);

              this.toastr.show(
                "The file has been uploaded successfully.",
                "success"
              );

            },
            error: (err) => {

              const message = err?.error?.error || "Failed to upload file";
              this.toastr.show(message, "error");

            }
          });

      };

      reader.readAsDataURL(file);

    });

  }
  onEditStaticFile(file: any): void {

    this.editingStaticFile = file;
    this.editingStaticFileContent = file.content;
    this.originalStaticFileContent = file.content;

  }
  hasStaticChanges(): boolean {
    return this.editingStaticFileContent !== this.originalStaticFileContent;
  }

  saveStaticFile(): void {

    if (!this.editingStaticFile || !this.templateMetadata) return;

    const payload: TemplateStaticFileRequest = {
      t_filename: this.templateMetadata.filename,
      filename: this.editingStaticFile.filename,
      content: this.editingStaticFileContent
    };

    this.api.updateTemplateStaticFile(payload).subscribe({

      next: () => {

        this.editingStaticFile.content = this.editingStaticFileContent;

        this.editingStaticFile = null;
        this.editingStaticFileContent = '';

        console.log('Static file updated');
        this.toastr.show("The file has been updated successfully.", "success");

      },

      error: (err) => {
        const message = err?.error?.error || "Failed to upload static file";
        this.toastr.show(message, "error");
      }

    });

  }

  onDeleteStaticFile(file: TemplateStaticFile): void {

    if (!this.templateMetadata) return;

    const payload: TemplateStaticFileRequest = {
      t_filename: this.templateMetadata.filename,
      filename: file.filename,
      content: ""
    };

    this.api.deleteTemplateStaticFile(payload)
      .subscribe({

        next: () => {

          console.log('Static file deleted:', file.filename);

          this.loadStaticFiles(this.templateMetadata!.filename);
          this.toastr.show("The file has been deleted successfully.", "success");
          this.router.navigate([
            '/templates',
          ]);
        },

        error: (err) => {
          const message = err?.error?.error || "Failed to delete file";
          this.toastr.show(message, "error");
        }

      });

  }

  getEditorLanguage(filename: string) {

    if (!filename) return this.language_html;

    if (filename.endsWith('.html')) return html();
    if (filename.endsWith('.css')) return css();
    if (filename.endsWith('.js')) return javascript();
    if (filename.endsWith('.yaml') || filename.endsWith('.yml')) return yaml();

    return html();
  }
}