import { Injectable } from '@angular/core';
import { CanActivate, Router, RouterStateSnapshot, ActivatedRouteSnapshot, UrlTree } from '@angular/router';
import { ApiService } from './services/api.service';
import { Observable } from 'rxjs';
import { AuthService } from './services/auth.service';

@Injectable({ providedIn: 'root' })
export class AuthGuard implements CanActivate {

    constructor(
        private router: Router,
        private api: ApiService, private auth: AuthService

    ) { }

    canActivate(): boolean | UrlTree {
        if (this.auth.isValidToken()) {
            return true;
        }

        return this.router.parseUrl('/login');
    }


}