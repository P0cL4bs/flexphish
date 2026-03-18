import { Injectable } from "@angular/core";
import { BehaviorSubject } from "rxjs";

@Injectable({ providedIn: 'root' })
export class SidebarService {
    private visible = new BehaviorSubject<boolean>(true);
    visible$ = this.visible.asObservable();

    toggle() {
        this.visible.next(!this.visible.value);
    }

    hide() {
        this.visible.next(false);
    }

    show() {
        this.visible.next(true);
    }
    isHide(): boolean {
        return this.visible.value
    }
}