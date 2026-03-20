import { CommonModule } from '@angular/common'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { html } from '@codemirror/lang-html'
import { CodeEditor } from '@acrodata/code-editor'
import { EmailTemplate, EmailTemplatePayload } from 'src/app/models/email-template.model'
import { ApiService } from 'src/app/services/api.service'
import { ToastService } from 'src/app/services/toast.service'
import { LucideAngularModule } from "lucide-angular";

type EmailTemplateForm = {
  name: string
  is_global: boolean
  subject: string
  body: string
}

type TargetPreviewData = {
  firstName: string
  lastName: string
  email: string
  position: string
  url: string
}

@Component({
  selector: 'app-email-templates-view',
  standalone: true,
  imports: [CommonModule, FormsModule, CodeEditor, LucideAngularModule],
  templateUrl: './email-templates-view.html',
  styleUrl: './email-templates-view.css'
})
export class EmailTemplatesView implements OnInit {
  language_html = html()
  placeholders = [
    { label: 'First Name', token: '{{FirstName}}' },
    { label: 'Last Name', token: '{{LastName}}' },
    { label: 'Email', token: '{{Email}}' },
    { label: 'Position', token: '{{Position}}' },
    { label: 'URL', token: '{{URL}}' }
  ]

  get bodyPlaceholders() {
    return this.placeholders
  }

  get subjectPlaceholders() {
    return this.placeholders.filter((p) => p.token !== '{{URL}}')
  }

  previewData: TargetPreviewData = {
    firstName: 'Ana',
    lastName: 'Silva',
    email: 'ana.silva@company.com',
    position: 'Manager',
    url: 'https://campaign.example.com/login'
  }

  templates: EmailTemplate[] = []
  filteredTemplates: EmailTemplate[] = []
  search = ''
  loading = false
  creating = false
  saving = false

  form: EmailTemplateForm = this.createEmptyForm()
  editingTemplate: EmailTemplate | null = null

  @ViewChild('createDialog') createDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('editDialog') editDialog!: ElementRef<HTMLDialogElement>

  constructor(private api: ApiService, private toast: ToastService) { }

  ngOnInit(): void {
    this.loadTemplates()
  }

  loadTemplates() {
    this.loading = true
    this.api.getEmailTemplates().subscribe({
      next: (templates) => {
        this.templates = templates ?? []
        this.applyFilter()
        this.loading = false
      },
      error: (err) => {
        this.loading = false
        this.toast.show(this.extractError(err, 'Failed to load email templates'), 'error')
      }
    })
  }

  applyFilter() {
    const term = this.search.trim().toLowerCase()
    if (!term) {
      this.filteredTemplates = [...this.templates]
      return
    }

    this.filteredTemplates = this.templates.filter((t) => {
      return (
        t.name.toLowerCase().includes(term) ||
        t.subject.toLowerCase().includes(term) ||
        t.body.toLowerCase().includes(term)
      )
    })
  }

  openCreateDialog() {
    this.form = this.createEmptyForm()
    this.createDialog?.nativeElement.showModal()
  }

  closeCreateDialog() {
    this.createDialog?.nativeElement.close()
  }

  openEditDialog(template: EmailTemplate) {
    this.editingTemplate = template
    this.form = {
      name: template.name || '',
      is_global: !!template.is_global,
      subject: template.subject || '',
      body: template.body || ''
    }
    this.editDialog?.nativeElement.showModal()
  }

  closeEditDialog() {
    this.editDialog?.nativeElement.close()
    this.editingTemplate = null
  }

  createTemplate() {
    const payload = this.buildPayload()
    if (!payload) {
      return
    }

    this.creating = true
    this.api.createEmailTemplate(payload).subscribe({
      next: () => {
        this.creating = false
        this.closeCreateDialog()
        this.toast.show('Email template created', 'success')
        this.loadTemplates()
      },
      error: (err) => {
        this.creating = false
        this.toast.show(this.extractError(err, 'Failed to create email template'), 'error')
      }
    })
  }

  updateTemplate() {
    if (!this.editingTemplate) {
      return
    }

    const payload = this.buildPayload()
    if (!payload) {
      return
    }

    this.saving = true
    this.api.updateEmailTemplate(this.editingTemplate.id, payload).subscribe({
      next: () => {
        this.saving = false
        this.closeEditDialog()
        this.toast.show('Email template updated', 'success')
        this.loadTemplates()
      },
      error: (err) => {
        this.saving = false
        this.toast.show(this.extractError(err, 'Failed to update email template'), 'error')
      }
    })
  }

  deleteTemplate(template: EmailTemplate) {
    const confirmed = window.confirm(`Delete email template "${template.name}"?`)
    if (!confirmed) {
      return
    }

    this.api.deleteEmailTemplate(template.id).subscribe({
      next: () => {
        this.toast.show('Email template deleted', 'success')
        this.loadTemplates()
      },
      error: (err) => {
        this.toast.show(this.extractError(err, 'Failed to delete email template'), 'error')
      }
    })
  }

  insertIntoBody(token: string) {
    const current = this.form.body || ''
    this.form.body = `${current}${current ? '\n' : ''}${token}`
  }

  insertIntoSubject(token: string) {
    const current = this.form.subject || ''
    this.form.subject = `${current}${current ? ' ' : ''}${token}`
  }

  get renderedSubject(): string {
    return this.applyPlaceholders(this.form.subject || '')
  }

  get renderedBody(): string {
    return this.applyPlaceholders(this.form.body || '')
  }

  private createEmptyForm(): EmailTemplateForm {
    return {
      name: '',
      is_global: false,
      subject: '',
      body: '<h2>Hello</h2><p>This is your email template.</p>'
    }
  }

  private buildPayload(): EmailTemplatePayload | null {
    const name = (this.form.name || '').trim()
    const subject = (this.form.subject || '').trim()
    const body = (this.form.body || '').trim()

    if (!name || !subject || !body) {
      this.toast.show('Name, subject and body are required', 'warning')
      return null
    }

    return {
      name,
      is_global: !!this.form.is_global,
      subject,
      body
    }
  }

  private applyPlaceholders(content: string): string {
    if (!content) {
      return ''
    }

    const replacements: Record<string, string> = {
      '{{FirstName}}': this.previewData.firstName || '',
      '{{LastName}}': this.previewData.lastName || '',
      '{{Email}}': this.previewData.email || '',
      '{{Position}}': this.previewData.position || '',
      '{{URL}}': this.previewData.url || ''
    }

    let rendered = content
    Object.entries(replacements).forEach(([key, value]) => {
      rendered = rendered.split(key).join(value)
    })

    return rendered
  }

  private extractError(err: any, fallback: string): string {
    if (typeof err?.error === 'string' && err.error.trim()) {
      return err.error
    }
    if (typeof err?.error?.error === 'string' && err.error.error.trim()) {
      return err.error.error
    }
    if (typeof err?.message === 'string' && err.message.trim()) {
      return err.message
    }
    return fallback
  }
}
