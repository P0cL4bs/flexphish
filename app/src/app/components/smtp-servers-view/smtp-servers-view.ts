import { CommonModule } from '@angular/common'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { SMTPProfile, SMTPProfilePayload, SMTPTestPayload } from 'src/app/models/smtp.model'
import { ApiService } from 'src/app/services/api.service'
import { ToastService } from 'src/app/services/toast.service'
import { LucideAngularModule } from "lucide-angular";

type SMTPForm = {
  name: string
  is_global: boolean
  host: string
  port: number | null
  username: string
  password: string
  from_name: string
  from_email: string
  test_email: string
  is_active: boolean
}

@Component({
  selector: 'app-smtp-servers-view',
  standalone: true,
  imports: [CommonModule, FormsModule, LucideAngularModule],
  templateUrl: './smtp-servers-view.html',
  styleUrl: './smtp-servers-view.css'
})
export class SMTPServersView implements OnInit {
  profiles: SMTPProfile[] = []
  filteredProfiles: SMTPProfile[] = []
  search = ''

  loading = false
  creating = false
  saving = false
  testing = false

  editingProfile: SMTPProfile | null = null
  smtpValidated = false

  form: SMTPForm = this.createEmptyForm()

  @ViewChild('createDialog') createDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('editDialog') editDialog!: ElementRef<HTMLDialogElement>

  constructor(private api: ApiService, private toast: ToastService) { }

  ngOnInit(): void {
    this.loadProfiles()
  }

  loadProfiles() {
    this.loading = true

    this.api.getSMTPProfiles().subscribe({
      next: (profiles) => {
        this.profiles = profiles ?? []
        this.applyFilter()
        this.loading = false
      },
      error: (err) => {
        this.loading = false
        this.toast.show(this.extractError(err, 'Failed to load SMTP servers'), 'error')
      }
    })
  }

  applyFilter() {
    const term = this.search.trim().toLowerCase()
    if (!term) {
      this.filteredProfiles = [...this.profiles]
      return
    }

    this.filteredProfiles = this.profiles.filter((p) => {
      return (
        p.name.toLowerCase().includes(term) ||
        p.host.toLowerCase().includes(term) ||
        p.username.toLowerCase().includes(term) ||
        (p.from_email || '').toLowerCase().includes(term)
      )
    })
  }

  openCreateDialog() {
    this.form = this.createEmptyForm()
    this.smtpValidated = false
    this.createDialog?.nativeElement.showModal()
  }

  closeCreateDialog() {
    this.createDialog?.nativeElement.close()
  }

  openEditDialog(profile: SMTPProfile) {
    this.editingProfile = profile
    this.form = {
      name: profile.name || '',
      is_global: !!profile.is_global,
      host: profile.host || '',
      port: profile.port || null,
      username: profile.username || '',
      password: '',
      from_name: profile.from_name || '',
      from_email: profile.from_email || '',
      test_email: '',
      is_active: !!profile.is_active
    }
    this.smtpValidated = false
    this.editDialog?.nativeElement.showModal()
  }

  closeEditDialog() {
    this.editDialog?.nativeElement.close()
    this.editingProfile = null
  }

  createProfile() {
    const payload = this.buildPayload()
    if (!payload) {
      return
    }
    if (!payload.password) {
      this.toast.show('Password is required for create', 'warning')
      return
    }
    if (!this.smtpValidated) {
      this.toast.show('Please test SMTP connection before creating', 'warning')
      return
    }

    this.creating = true
    this.api.createSMTPProfile(payload).subscribe({
      next: () => {
        this.creating = false
        this.closeCreateDialog()
        this.toast.show('SMTP server created', 'success')
        this.loadProfiles()
      },
      error: (err) => {
        this.creating = false
        this.toast.show(this.extractError(err, 'Failed to create SMTP server'), 'error')
      }
    })
  }

  updateProfile() {
    if (!this.editingProfile) {
      return
    }

    const payload = this.buildPayload()
    if (!payload) {
      return
    }

    this.saving = true
    this.api.updateSMTPProfile(this.editingProfile.id, payload).subscribe({
      next: () => {
        this.saving = false
        this.closeEditDialog()
        this.toast.show('SMTP server updated', 'success')
        this.loadProfiles()
      },
      error: (err) => {
        this.saving = false
        this.toast.show(this.extractError(err, 'Failed to update SMTP server'), 'error')
      }
    })
  }

  deleteProfile(profile: SMTPProfile) {
    const confirmed = window.confirm(`Delete SMTP server "${profile.name}"?`)
    if (!confirmed) {
      return
    }

    this.api.deleteSMTPProfile(profile.id).subscribe({
      next: () => {
        this.toast.show('SMTP server deleted', 'success')
        this.loadProfiles()
      },
      error: (err) => {
        this.toast.show(this.extractError(err, 'Failed to delete SMTP server'), 'error')
      }
    })
  }

  private createEmptyForm(): SMTPForm {
    return {
      name: '',
      is_global: false,
      host: '',
      port: 587,
      username: '',
      password: '',
      from_name: '',
      from_email: '',
      test_email: '',
      is_active: true
    }
  }

  private buildPayload(): SMTPProfilePayload | null {
    const name = (this.form.name || '').trim()
    const host = (this.form.host || '').trim()
    const username = (this.form.username || '').trim()
    const password = (this.form.password || '').trim()
    const from_name = (this.form.from_name || '').trim()
    const from_email = (this.form.from_email || '').trim()
    const port = Number(this.form.port)

    if (!name || !host || !username || !Number.isFinite(port) || port <= 0) {
      this.toast.show('Please fill required fields: name, host, port and username', 'warning')
      return null
    }

    return {
      name,
      is_global: !!this.form.is_global,
      host,
      port,
      username,
      password,
      from_name,
      from_email,
      is_active: !!this.form.is_active
    }
  }

  testConnection() {
    const host = (this.form.host || '').trim()
    const username = (this.form.username || '').trim()
    const password = (this.form.password || '').trim()
    const from_name = (this.form.from_name || '').trim()
    const from_email = (this.form.from_email || '').trim()
    const test_email = (this.form.test_email || '').trim()
    const port = Number(this.form.port)

    if (!host || !username || !password || !test_email || !Number.isFinite(port) || port <= 0) {
      this.toast.show('Fill host, port, username, password and test email', 'warning')
      return
    }

    const payload: SMTPTestPayload = {
      name: (this.form.name || '').trim(),
      host,
      port,
      username,
      password,
      from_name,
      from_email,
      test_email
    }

    this.testing = true
    this.smtpValidated = false
    this.api.testSMTPProfile(payload).subscribe({
      next: (res) => {
        this.testing = false
        this.smtpValidated = true
        this.toast.show(res?.message || 'Test email sent successfully', 'success')
      },
      error: (err) => {
        this.testing = false
        this.smtpValidated = false
        this.toast.show(this.extractError(err, 'SMTP test failed'), 'error')
      }
    })
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
