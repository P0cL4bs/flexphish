import { CommonModule } from '@angular/common'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { FormsModule } from '@angular/forms'
import { firstValueFrom } from 'rxjs'
import { Group, GroupTarget, GroupTargetPayload } from 'src/app/models/group.model'
import { ApiService } from 'src/app/services/api.service'
import { ToastService } from 'src/app/services/toast.service'
import { LucideAngularModule } from "lucide-angular";

type TargetForm = {
  first_name: string
  last_name: string
  email: string
  position: string
}

type ImportedTargetRow = TargetForm & {
  sourceLine: number
}

@Component({
  selector: 'app-groups-view',
  standalone: true,
  imports: [CommonModule, FormsModule, LucideAngularModule],
  templateUrl: './groups-view.html',
  styleUrl: './groups-view.css'
})
export class GroupsView implements OnInit {
  groups: Group[] = []
  filteredGroups: Group[] = []
  targets: GroupTarget[] = []
  filteredTargets: GroupTarget[] = []

  selectedGroup: Group | null = null

  groupsSearch = ''
  targetsSearch = ''

  loadingGroups = false
  loadingTargets = false
  creatingGroup = false
  savingGroup = false
  creatingTarget = false
  savingTarget = false
  importingTargets = false

  groupForm = {
    name: '',
    is_global: false
  }

  createGroupTargets: TargetForm[] = []

  targetForm: TargetForm = {
    first_name: '',
    last_name: '',
    email: '',
    position: ''
  }

  editingGroup: Group | null = null
  editingTarget: GroupTarget | null = null

  @ViewChild('createGroupDialog') createGroupDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('editGroupDialog') editGroupDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('createTargetDialog') createTargetDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('editTargetDialog') editTargetDialog!: ElementRef<HTMLDialogElement>
  @ViewChild('targetsImportInput') targetsImportInput!: ElementRef<HTMLInputElement>

  constructor(private api: ApiService, private toast: ToastService) { }

  ngOnInit(): void {
    this.loadGroups()
  }

  loadGroups(preserveSelected = true, preferGroupId?: number) {
    this.loadingGroups = true
    const selectedId = preserveSelected ? this.selectedGroup?.id : null

    this.api.getGroups().subscribe({
      next: (groups) => {
        this.groups = groups ?? []
        this.applyGroupFilter()
        this.loadingGroups = false

        if (this.groups.length === 0) {
          this.selectedGroup = null
          this.targets = []
          this.filteredTargets = []
          return
        }

        const nextGroup = preferGroupId
          ? this.groups.find((g) => g.id === preferGroupId) || this.groups[0]
          : selectedId
            ? this.groups.find((g) => g.id === selectedId) || this.groups[0]
            : this.groups[0]

        this.selectGroup(nextGroup)
      },
      error: (err) => {
        this.loadingGroups = false
        this.toast.show(this.extractError(err, 'Failed to load groups'), 'error')
      }
    })
  }

  loadTargets(groupId: number) {
    this.loadingTargets = true

    this.api.getGroupTargets(groupId).subscribe({
      next: (targets) => {
        this.targets = targets ?? []
        this.applyTargetFilter()
        this.loadingTargets = false
      },
      error: (err) => {
        this.loadingTargets = false
        this.toast.show(this.extractError(err, 'Failed to load targets'), 'error')
      }
    })
  }

  selectGroup(group: Group) {
    this.selectedGroup = group
    this.targetsSearch = ''
    this.loadTargets(group.id)
  }

  applyGroupFilter() {
    const term = this.groupsSearch.trim().toLowerCase()
    if (!term) {
      this.filteredGroups = [...this.groups]
      return
    }

    this.filteredGroups = this.groups.filter((g) => {
      return g.name.toLowerCase().includes(term)
    })
  }

  applyTargetFilter() {
    const term = this.targetsSearch.trim().toLowerCase()
    if (!term) {
      this.filteredTargets = [...this.targets]
      return
    }

    this.filteredTargets = this.targets.filter((t) => {
      return (
        t.first_name?.toLowerCase().includes(term) ||
        t.last_name?.toLowerCase().includes(term) ||
        t.email?.toLowerCase().includes(term) ||
        t.position?.toLowerCase().includes(term)
      )
    })
  }

  openCreateGroupDialog() {
    this.groupForm = { name: '', is_global: false }
    this.createGroupTargets = []
    this.createGroupDialog?.nativeElement.showModal()
  }

  closeCreateGroupDialog() {
    this.createGroupDialog?.nativeElement.close()
  }

  openEditGroupDialog(group: Group, event?: MouseEvent) {
    event?.stopPropagation()
    this.editingGroup = group
    this.groupForm = {
      name: group.name,
      is_global: group.is_global
    }
    this.editGroupDialog?.nativeElement.showModal()
  }

  closeEditGroupDialog() {
    this.editGroupDialog?.nativeElement.close()
    this.editingGroup = null
  }

  addInitialTarget() {
    this.createGroupTargets.push({
      first_name: '',
      last_name: '',
      email: '',
      position: ''
    })
  }

  removeInitialTarget(index: number) {
    this.createGroupTargets.splice(index, 1)
  }

  createGroup() {
    const name = this.groupForm.name.trim()
    if (!name) {
      this.toast.show('Group name is required', 'warning')
      return
    }

    const payloadTargets: GroupTargetPayload[] = []
    for (const row of this.createGroupTargets) {
      const candidate = this.normalizeTargetForm(row)
      const hasAnyField = !!(candidate.first_name || candidate.last_name || candidate.email || candidate.position)
      if (!hasAnyField) {
        continue
      }
      if (!candidate.email) {
        this.toast.show('Initial target email is required', 'warning')
        return
      }
      payloadTargets.push(candidate)
    }

    this.creatingGroup = true
    this.api.createGroup({
      name,
      is_global: this.groupForm.is_global,
      targets: payloadTargets
    }).subscribe({
      next: (created) => {
        this.creatingGroup = false
        this.closeCreateGroupDialog()
        this.toast.show('Group created successfully', 'success')
        this.loadGroups(false, created?.id)
      },
      error: (err) => {
        this.creatingGroup = false
        this.toast.show(this.extractError(err, 'Failed to create group'), 'error')
      }
    })
  }

  updateGroup() {
    if (!this.editingGroup) {
      return
    }
    const name = this.groupForm.name.trim()
    if (!name) {
      this.toast.show('Group name is required', 'warning')
      return
    }

    this.savingGroup = true
    this.api.updateGroup(this.editingGroup.id, {
      name,
      is_global: this.groupForm.is_global
    }).subscribe({
      next: () => {
        this.savingGroup = false
        this.closeEditGroupDialog()
        this.toast.show('Group updated', 'success')
        this.loadGroups(true)
      },
      error: (err) => {
        this.savingGroup = false
        this.toast.show(this.extractError(err, 'Failed to update group'), 'error')
      }
    })
  }

  deleteGroup(group: Group, event?: MouseEvent) {
    event?.stopPropagation()
    const confirmed = window.confirm(`Delete group "${group.name}"?`)
    if (!confirmed) {
      return
    }

    this.api.deleteGroup(group.id).subscribe({
      next: () => {
        if (this.selectedGroup?.id === group.id) {
          this.selectedGroup = null
          this.targets = []
          this.filteredTargets = []
        }
        this.toast.show('Group deleted', 'success')
        this.loadGroups(false)
      },
      error: (err) => {
        this.toast.show(this.extractError(err, 'Failed to delete group'), 'error')
      }
    })
  }

  openCreateTargetDialog() {
    if (!this.selectedGroup) {
      return
    }
    this.targetForm = {
      first_name: '',
      last_name: '',
      email: '',
      position: ''
    }
    this.createTargetDialog?.nativeElement.showModal()
  }

  triggerTargetImport() {
    if (!this.selectedGroup || this.importingTargets) {
      return
    }
    this.targetsImportInput?.nativeElement.click()
  }

  async onTargetsFileSelected(event: Event) {
    if (!this.selectedGroup) {
      return
    }

    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) {
      return
    }

    try {
      const importedRows = await this.parseTargetsFile(file)
      if (importedRows.length === 0) {
        this.toast.show('No targets found in the file', 'warning')
        return
      }

      const validation = this.validateImportedTargets(importedRows)
      if (validation.errors.length > 0) {
        this.toast.show(validation.errors[0], 'error', 5000)
        return
      }

      const confirmed = window.confirm(`Import ${validation.targets.length} target(s) to "${this.selectedGroup.name}"?`)
      if (!confirmed) {
        return
      }

      await this.importTargets(validation.targets)
    } catch (err: any) {
      this.toast.show(this.extractError(err, 'Failed to import file'), 'error', 5000)
    } finally {
      input.value = ''
    }
  }

  closeCreateTargetDialog() {
    this.createTargetDialog?.nativeElement.close()
  }

  openEditTargetDialog(target: GroupTarget) {
    this.editingTarget = target
    this.targetForm = {
      first_name: target.first_name || '',
      last_name: target.last_name || '',
      email: target.email || '',
      position: target.position || ''
    }
    this.editTargetDialog?.nativeElement.showModal()
  }

  closeEditTargetDialog() {
    this.editTargetDialog?.nativeElement.close()
    this.editingTarget = null
  }

  createTarget() {
    if (!this.selectedGroup) {
      return
    }
    const payload = this.normalizeTargetForm(this.targetForm)
    if (!payload.email) {
      this.toast.show('Email is required', 'warning')
      return
    }

    this.creatingTarget = true
    this.api.createGroupTarget(this.selectedGroup.id, payload).subscribe({
      next: () => {
        this.creatingTarget = false
        this.closeCreateTargetDialog()
        this.toast.show('Target created', 'success')
        this.loadTargets(this.selectedGroup!.id)
        this.loadGroups(true)
      },
      error: (err) => {
        this.creatingTarget = false
        this.toast.show(this.extractError(err, 'Failed to create target'), 'error')
      }
    })
  }

  updateTarget() {
    if (!this.selectedGroup || !this.editingTarget) {
      return
    }

    const payload = this.normalizeTargetForm(this.targetForm)
    if (!payload.email) {
      this.toast.show('Email is required', 'warning')
      return
    }

    this.savingTarget = true
    this.api.updateGroupTarget(this.selectedGroup.id, this.editingTarget.id, payload).subscribe({
      next: () => {
        this.savingTarget = false
        this.closeEditTargetDialog()
        this.toast.show('Target updated', 'success')
        this.loadTargets(this.selectedGroup!.id)
      },
      error: (err) => {
        this.savingTarget = false
        this.toast.show(this.extractError(err, 'Failed to update target'), 'error')
      }
    })
  }

  deleteTarget(target: GroupTarget) {
    if (!this.selectedGroup) {
      return
    }
    const confirmed = window.confirm(`Delete target "${target.email}"?`)
    if (!confirmed) {
      return
    }

    this.api.deleteGroupTarget(this.selectedGroup.id, target.id).subscribe({
      next: () => {
        this.toast.show('Target deleted', 'success')
        this.loadTargets(this.selectedGroup!.id)
        this.loadGroups(true)
      },
      error: (err) => {
        this.toast.show(this.extractError(err, 'Failed to delete target'), 'error')
      }
    })
  }

  getTargetCount(group: Group) {
    return group.targets?.length || 0
  }

  getGroupInitials(name: string) {
    const cleanName = (name || '').trim()
    if (!cleanName) {
      return 'GR'
    }

    const parts = cleanName.split(/\s+/).filter(Boolean)
    if (parts.length === 1) {
      return parts[0].slice(0, 2).toUpperCase()
    }

    return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
  }

  private normalizeTargetForm(input: TargetForm): GroupTargetPayload {
    return {
      first_name: (input.first_name || '').trim(),
      last_name: (input.last_name || '').trim(),
      email: (input.email || '').trim(),
      position: (input.position || '').trim()
    }
  }

  private async importTargets(targets: ImportedTargetRow[]) {
    if (!this.selectedGroup) {
      return
    }

    this.importingTargets = true
    let importedCount = 0
    const failures: string[] = []

    for (const row of targets) {
      const payload = this.normalizeTargetForm(row)
      try {
        await firstValueFrom(this.api.createGroupTarget(this.selectedGroup.id, payload))
        importedCount += 1
      } catch (err) {
        failures.push(`Line ${row.sourceLine}: ${this.extractError(err, 'Failed to create target')}`)
      }
    }

    this.importingTargets = false

    if (importedCount > 0) {
      this.loadTargets(this.selectedGroup.id)
      this.loadGroups(true)
    }

    if (failures.length === 0) {
      this.toast.show(`Imported ${importedCount} target(s) successfully`, 'success')
      return
    }

    this.toast.show(`Imported ${importedCount} target(s), ${failures.length} failed`, 'warning', 5000)
    this.toast.show(failures[0], 'error', 6000)
  }

  private async parseTargetsFile(file: File): Promise<ImportedTargetRow[]> {
    const content = await file.text()
    const fileName = file.name.toLowerCase()

    if (fileName.endsWith('.csv')) {
      return this.parseCsvTargets(content)
    }
    if (fileName.endsWith('.xml')) {
      return this.parseXmlTargets(content)
    }

    throw new Error('Unsupported file type. Use .csv or .xml')
  }

  private parseCsvTargets(content: string): ImportedTargetRow[] {
    const rows = this.parseCsvLines(content)
    if (rows.length === 0) {
      throw new Error('CSV file is empty')
    }

    const headers = rows[0].map((header) => this.normalizeImportFieldName(header))
    const required = ['first_name', 'last_name', 'email', 'position']
    const missing = required.filter((field) => !headers.includes(field))
    if (missing.length > 0) {
      throw new Error(`CSV missing required columns: ${missing.join(', ')}`)
    }

    const indexByField: Record<string, number> = {}
    headers.forEach((header, index) => {
      if (required.includes(header) && indexByField[header] === undefined) {
        indexByField[header] = index
      }
    })

    const imported: ImportedTargetRow[] = []
    for (let i = 1; i < rows.length; i += 1) {
      const values = rows[i]
      const first_name = (values[indexByField['first_name']] || '').trim()
      const last_name = (values[indexByField['last_name']] || '').trim()
      const email = (values[indexByField['email']] || '').trim()
      const position = (values[indexByField['position']] || '').trim()

      if (!(first_name || last_name || email || position)) {
        continue
      }

      imported.push({
        first_name,
        last_name,
        email,
        position,
        sourceLine: i + 1
      })
    }

    return imported
  }

  private parseXmlTargets(content: string): ImportedTargetRow[] {
    const parser = new DOMParser()
    const doc = parser.parseFromString(content, 'application/xml')

    if (doc.getElementsByTagName('parsererror').length > 0) {
      throw new Error('Invalid XML file')
    }

    const targetNodes = Array.from(doc.getElementsByTagName('target'))
    if (targetNodes.length === 0) {
      throw new Error('XML must include at least one <target> node')
    }

    return targetNodes.map((node, index) => {
      const missingFields: string[] = []
      if (!this.hasXmlField(node, ['first_name', 'firstName', 'firstname'])) {
        missingFields.push('first_name')
      }
      if (!this.hasXmlField(node, ['last_name', 'lastName', 'lastname'])) {
        missingFields.push('last_name')
      }
      if (!this.hasXmlField(node, ['email'])) {
        missingFields.push('email')
      }
      if (!this.hasXmlField(node, ['position'])) {
        missingFields.push('position')
      }
      if (missingFields.length > 0) {
        throw new Error(`XML target #${index + 1} is missing fields: ${missingFields.join(', ')}`)
      }

      return {
        first_name: this.readXmlField(node, ['first_name', 'firstName', 'firstname']),
        last_name: this.readXmlField(node, ['last_name', 'lastName', 'lastname']),
        email: this.readXmlField(node, ['email']),
        position: this.readXmlField(node, ['position']),
        sourceLine: index + 1
      }
    })
  }

  private validateImportedTargets(rows: ImportedTargetRow[]): { targets: ImportedTargetRow[], errors: string[] } {
    const errors: string[] = []
    const emails = new Set<string>()

    for (const row of rows) {
      const normalized = this.normalizeTargetForm(row)
      const linePrefix = `Line ${row.sourceLine}`

      if (!normalized.email) {
        errors.push(`${linePrefix}: email is required`)
        continue
      }

      if (!this.isValidEmail(normalized.email)) {
        errors.push(`${linePrefix}: invalid email "${normalized.email}"`)
        continue
      }

      const lowerEmail = normalized.email.toLowerCase()
      if (emails.has(lowerEmail)) {
        errors.push(`${linePrefix}: duplicate email "${normalized.email}" in file`)
        continue
      }
      emails.add(lowerEmail)
    }

    return { targets: rows, errors }
  }

  private readXmlField(node: Element, fieldNames: string[]): string {
    for (const fieldName of fieldNames) {
      const value = node.getElementsByTagName(fieldName)[0]?.textContent
      if (typeof value === 'string') {
        return value.trim()
      }
    }
    return ''
  }

  private hasXmlField(node: Element, fieldNames: string[]): boolean {
    return fieldNames.some((fieldName) => node.getElementsByTagName(fieldName).length > 0)
  }

  private normalizeImportFieldName(field: string): string {
    const normalized = (field || '')
      .trim()
      .toLowerCase()
      .replace(/[\s-]+/g, '_')

    switch (normalized) {
      case 'firstname':
      case 'first_name':
        return 'first_name'
      case 'lastname':
      case 'last_name':
        return 'last_name'
      case 'email':
        return 'email'
      case 'position':
        return 'position'
      default:
        return normalized
    }
  }

  private parseCsvLines(content: string): string[][] {
    const lines: string[][] = []
    let currentLine: string[] = []
    let currentValue = ''
    let inQuotes = false

    for (let i = 0; i < content.length; i += 1) {
      const char = content[i]
      const next = content[i + 1]

      if (char === '"') {
        if (inQuotes && next === '"') {
          currentValue += '"'
          i += 1
        } else {
          inQuotes = !inQuotes
        }
        continue
      }

      if (char === ',' && !inQuotes) {
        currentLine.push(currentValue)
        currentValue = ''
        continue
      }

      if ((char === '\n' || char === '\r') && !inQuotes) {
        if (char === '\r' && next === '\n') {
          i += 1
        }
        currentLine.push(currentValue)
        lines.push(currentLine)
        currentLine = []
        currentValue = ''
        continue
      }

      currentValue += char
    }

    if (currentValue.length > 0 || currentLine.length > 0) {
      currentLine.push(currentValue)
      lines.push(currentLine)
    }

    return lines.filter((line, index) => {
      if (index === 0) {
        return true
      }
      return line.some((part) => part.trim() !== '')
    })
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
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
