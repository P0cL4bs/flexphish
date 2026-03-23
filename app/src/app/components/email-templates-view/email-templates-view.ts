import { CommonModule } from '@angular/common'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser'
import { html } from '@codemirror/lang-html'
import { CodeEditor } from '@acrodata/code-editor'
import { EmailTemplate, EmailTemplateAttachment, EmailTemplatePayload } from 'src/app/models/email-template.model'
import { ApiService } from 'src/app/services/api.service'
import { ToastService } from 'src/app/services/toast.service'
import { LucideAngularModule } from "lucide-angular";

type EmailTemplateForm = {
  name: string
  category: string
  is_global: boolean
  track_opens: boolean
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
  private previewSrcCache = new Map<string, SafeResourceUrl>()
  search = ''
  loading = false
  creating = false
  saving = false
  loadingAttachments = false
  uploadingAttachment = false
  attachments: EmailTemplateAttachment[] = []

  form: EmailTemplateForm = this.createEmptyForm()
  editingTemplate: EmailTemplate | null = null

  @ViewChild('createDialog') createDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('editDialog') editDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('createBodyEditor') createBodyEditor?: CodeEditor
  @ViewChild('editBodyEditor') editBodyEditor?: CodeEditor
  @ViewChild('createSubjectInput') createSubjectInput?: ElementRef<HTMLInputElement>
  @ViewChild('editSubjectInput') editSubjectInput?: ElementRef<HTMLInputElement>

  constructor(
    private api: ApiService,
    private toast: ToastService,
    private sanitizer: DomSanitizer
  ) { }

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
      category: template.category || '',
      is_global: !!template.is_global,
      track_opens: template.track_opens ?? true,
      subject: template.subject || '',
      body: template.body || ''
    }
    this.loadAttachments(template.id)
    this.editDialog?.nativeElement.showModal()
  }

  closeEditDialog() {
    this.editDialog?.nativeElement.close()
    this.editingTemplate = null
    this.attachments = []
    this.loadingAttachments = false
    this.uploadingAttachment = false
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

  loadAttachments(templateId: number) {
    this.loadingAttachments = true
    this.api.getEmailTemplateAttachments(templateId).subscribe({
      next: (attachments) => {
        this.attachments = attachments ?? []
        this.loadingAttachments = false
      },
      error: (err) => {
        this.loadingAttachments = false
        this.toast.show(this.extractError(err, 'Failed to load attachments'), 'error')
      }
    })
  }

  onAttachmentSelected(event: Event) {
    if (!this.editingTemplate) {
      return
    }

    const input = event.target as HTMLInputElement
    const file = input?.files?.[0]
    if (!file) {
      return
    }

    this.uploadingAttachment = true
    this.api.uploadEmailTemplateAttachment(this.editingTemplate.id, file).subscribe({
      next: () => {
        this.uploadingAttachment = false
        input.value = ''
        this.toast.show('Attachment uploaded', 'success')
        this.loadAttachments(this.editingTemplate!.id)
      },
      error: (err) => {
        this.uploadingAttachment = false
        this.toast.show(this.extractError(err, 'Failed to upload attachment'), 'error')
      }
    })
  }

  deleteAttachment(attachment: EmailTemplateAttachment) {
    if (!this.editingTemplate) {
      return
    }

    const confirmed = window.confirm(`Delete attachment "${attachment.filename}"?`)
    if (!confirmed) {
      return
    }

    this.api.deleteEmailTemplateAttachment(this.editingTemplate.id, attachment.id).subscribe({
      next: () => {
        this.toast.show('Attachment deleted', 'success')
        this.loadAttachments(this.editingTemplate!.id)
      },
      error: (err) => {
        this.toast.show(this.extractError(err, 'Failed to delete attachment'), 'error')
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

  insertIntoBody(token: string, context: 'create' | 'edit', event?: Event) {
    event?.preventDefault()
    event?.stopPropagation()

    const editor = context === 'create' ? this.createBodyEditor : this.editBodyEditor
    const view = editor?.view
    if (view) {
      const selection = view.state.selection.main
      view.dispatch({
        changes: { from: selection.from, to: selection.to, insert: token },
        selection: { anchor: selection.from + token.length }
      })
      this.form = {
        ...this.form,
        body: view.state.doc.toString()
      }
      view.focus()
      return
    }

    const current = this.form.body || ''
    this.form = {
      ...this.form,
      body: `${current}${token}`
    }
  }

  insertIntoSubject(token: string, context: 'create' | 'edit', event?: Event) {
    event?.preventDefault()
    event?.stopPropagation()

    const inputRef = context === 'create' ? this.createSubjectInput : this.editSubjectInput
    const subjectInput = inputRef?.nativeElement

    if (subjectInput) {
      const current = this.form.subject || ''
      const from = subjectInput.selectionStart ?? current.length
      const to = subjectInput.selectionEnd ?? current.length
      const next = `${current.slice(0, from)}${token}${current.slice(to)}`
      const nextCursor = from + token.length

      this.form = {
        ...this.form,
        subject: next
      }

      requestAnimationFrame(() => {
        subjectInput.focus()
        subjectInput.setSelectionRange(nextCursor, nextCursor)
      })
      return
    }

    const current = this.form.subject || ''
    this.form = {
      ...this.form,
      subject: `${current}${token}`
    }
  }

  get renderedSubject(): string {
    return this.applyPlaceholders(this.form.subject || '')
  }

  get renderedBody(): string {
    return this.applyPlaceholders(this.form.body || '')
  }

  get renderedBodyDoc(): string {
    return this.buildPreviewDocument(this.renderedBody)
  }

  get renderedBodyPreviewSrc(): SafeResourceUrl {
    return this.toPreviewSrc(this.renderedBodyDoc)
  }

  renderTemplateSubject(template: EmailTemplate): string {
    return this.applyPlaceholders(template?.subject || '')
  }

  renderTemplateBody(template: EmailTemplate): string {
    return this.applyPlaceholders(template?.body || '')
  }

  renderTemplateBodyDoc(template: EmailTemplate): string {
    return this.buildPreviewDocument(this.renderTemplateBody(template))
  }

  renderTemplatePreviewSrc(template: EmailTemplate): SafeResourceUrl {
    return this.toPreviewSrc(this.renderTemplateBodyDoc(template))
  }

  private createEmptyForm(): EmailTemplateForm {
    return {
      name: '',
      category: '',
      is_global: false,
      track_opens: true,
      subject: '',
      body: '<h2>Hello</h2><p>This is your email template.</p>'
    }
  }

  private buildPayload(): EmailTemplatePayload | null {
    const name = (this.form.name || '').trim()
    const category = (this.form.category || '').trim()
    const subject = (this.form.subject || '').trim()
    const body = (this.form.body || '').trim()

    if (!name || !subject || !body) {
      this.toast.show('Name, subject and body are required', 'warning')
      return null
    }

    return {
      name,
      category,
      is_global: !!this.form.is_global,
      track_opens: !!this.form.track_opens,
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

  private buildPreviewDocument(content: string): string {
    const trimmed = (content || '').trim()
    const lower = trimmed.toLowerCase()

    const isFullDocument =
      lower.startsWith('<!doctype') ||
      lower.includes('<html') ||
      lower.includes('<head') ||
      lower.includes('<body')

    if (isFullDocument) {
      return trimmed
    }

    return `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      html, body { margin: 0; padding: 0; background: #ffffff; color: #111827; font-family: Arial, sans-serif; }
      body { padding: 12px; box-sizing: border-box; }
      * { box-sizing: border-box; }
      img { max-width: 100%; height: auto; }
      a { color: #2563eb; }
    </style>
  </head>
  <body>${trimmed}</body>
</html>`
  }

  private toPreviewSrc(htmlDocument: string): SafeResourceUrl {
    const document = htmlDocument || ''
    const cached = this.previewSrcCache.get(document)
    if (cached) {
      return cached
    }

    const encoded = encodeURIComponent(document)
    const safeUrl = this.sanitizer.bypassSecurityTrustResourceUrl(`data:text/html;charset=utf-8,${encoded}`)
    this.previewSrcCache.set(document, safeUrl)
    return safeUrl
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

  formatBytes(size?: number): string {
    if (!size || size <= 0) {
      return '0 B'
    }
    const units = ['B', 'KB', 'MB', 'GB']
    let value = size
    let unit = 0
    for (; value >= 1024 && unit < units.length - 1; unit++) {
      value /= 1024
    }
    const decimals = unit === 0 ? 0 : 1
    return `${value.toFixed(decimals)} ${units[unit]}`
  }
}
