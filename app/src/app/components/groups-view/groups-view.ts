import { CommonModule } from '@angular/common'
import { Component, ElementRef, OnInit, ViewChild } from '@angular/core'
import { FormsModule } from '@angular/forms'
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

  private normalizeTargetForm(input: TargetForm): GroupTargetPayload {
    return {
      first_name: (input.first_name || '').trim(),
      last_name: (input.last_name || '').trim(),
      email: (input.email || '').trim(),
      position: (input.position || '').trim()
    }
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
