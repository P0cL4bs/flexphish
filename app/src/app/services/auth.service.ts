import { Router, UrlTree } from "@angular/router"
import { ApiService } from "./api.service"
import { Injectable } from "@angular/core"
import { catchError, map, Observable, of } from "rxjs"

@Injectable({
    providedIn: 'root'
})
export class AuthService {

    constructor(private api: ApiService, private router: Router) { }

    validateSession(): Observable<boolean | UrlTree> {
        return this.api.validateToken().pipe(
            map(() => true),
            catchError(() => {
                this.api.creds.clear();
                return of(this.router.createUrlTree(['/login']));
            })
        )
    }

    logout() {
        this.api.creds.clear()
        this.router.navigate(['/login'])
    }

    isValidToken(): boolean {
        return this.api.creds.isValidToken()
    }

}