import { CodeEditor } from '@acrodata/code-editor';
import { CommonModule } from '@angular/common';
import { Component, ElementRef, ViewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { yaml } from '@codemirror/lang-yaml';
import { TemplateCreateRequest, TemplateHtmlFile, TemplateHtmlFileUploadRequest, TemplateStaticFile, TemplateStaticFileRequest } from 'src/app/models/template.model';
import { ApiService } from 'src/app/services/api.service';
import { HumanizeBytesPipe } from "../humanize-bytes.pipe";
import { ToastService } from 'src/app/services/toast.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-template-create-view',
  imports: [CommonModule,
    FormsModule, CodeEditor, HumanizeBytesPipe],
  templateUrl: './template-create-view.html',
  styleUrl: './template-create-view.css'
})
export class TemplateCreateView {
  @ViewChild('createTemplateDialog')
  dialog!: ElementRef<HTMLDialogElement>;
  filename = '';
  content = '';
  language_yaml = yaml();

  htmlFiles: File[] = [];
  staticFiles: File[] = [];

  loading = false;

  constructor(private api: ApiService, private toastr: ToastService, private router: Router) { }


  openModal() {

    if (this.dialog) {

      this.dialog.nativeElement.showModal();

    }

  }

  closeModal() {

    if (this.dialog) {

      this.dialog.nativeElement.close();

    }

  }

  onHtmlFiles(event: any) {

    const files = Array.from(event.target.files) as File[];

    this.htmlFiles = files;

  }

  onStaticFiles(event: any) {

    const files = Array.from(event.target.files) as File[];

    this.staticFiles = files;

  }

  createTemplate() {

    if (!this.filename.endsWith(".yaml")) {

      alert("Template filename must end with .yaml");

      return;

    }

    const request = {
      filename: this.filename,
      content: this.content
    };

    this.loading = true;

    this.api.createTemplate(request).subscribe({

      next: (res) => {

        const templateId = request.filename;

        console.log("Template created");
        this.toastr.show("The tempalte has been created successfully.", "success");

        this.uploadHtmlFiles(templateId);

      },

      error: (err) => {

        console.error("Template creation failed", err);

        this.loading = false;
        const message = err?.error?.error || "Template creation failed";
        this.toastr.show(message, "error");
      }

    });

  }

  uploadHtmlFiles(templateId: string) {

    if (this.htmlFiles.length === 0) {

      this.uploadStaticFiles();

      return;

    }

    let processed = 0;

    this.htmlFiles.forEach(file => {

      const reader = new FileReader();

      reader.onload = () => {

        const data: TemplateHtmlFileUploadRequest = {
          filename: file.name,
          t_filename: templateId,
          content: reader.result as string
        };

        this.api.uploadTemplateHtmlFile(data, templateId).subscribe({

          next: () => {

            processed++;

            if (processed === this.htmlFiles.length) {

              console.log("All HTML files uploaded");

              this.uploadStaticFiles();
              this.toastr.show("All  HTML files has been created successfully.", "success");

            }

          },

          error: (err) => {

            console.error("HTML upload error", err);

            this.loading = false;
            const message = err?.error?.error || "HTML upload error";
            this.toastr.show(message, "error");

          }

        });

      };

      reader.readAsText(file);

    });

  }

  uploadStaticFiles() {

    if (this.staticFiles.length === 0) {

      this.finish();

      return;

    }

    let processed = 0;

    this.staticFiles.forEach(file => {

      const reader = new FileReader();

      reader.onload = () => {

        const data: TemplateStaticFileRequest = {

          t_filename: this.filename,
          filename: file.name,
          content: reader.result as string

        };

        this.api.createTemplateStaticFile(data).subscribe({

          next: () => {

            processed++;

            if (processed === this.staticFiles.length) {

              console.log("All static files uploaded");

              this.finish();
              this.toastr.show("All Static files has been created successfully.", "success");
            }

          },

          error: (err) => {

            console.error("Static upload error", err);

            this.loading = false;
            const message = err?.error?.error || "Static upload error";
            this.toastr.show(message, "error");

          }

        });

      };

      reader.readAsDataURL(file);

    });

  }

  finish() {
    this.loading = false;
    this.closeModal()
    window.dispatchEvent(new Event('templates:reload'));
  }

}
