import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { ApiService } from 'src/app/services/api.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const api = inject(ApiService);
  const router = inject(Router);

  const token = api.creds.getToken();

  const authReq = token
    ? req.clone({
      setHeaders: { Authorization: `Bearer ${token}` }
    })
    : req;

  return next(authReq).pipe(
    catchError(err => {
      if (err.status === 401) {
        api.creds.clear();
        router.navigate(['/login']);
      }

      return throwError(() => err);
    })
  );
};