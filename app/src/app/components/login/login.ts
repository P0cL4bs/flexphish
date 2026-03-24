import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/module.d-yNBsZ8gb';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { ApiService } from 'src/app/services/api.service';
import { environment } from 'src/environments/environment';

@Component({
  selector: 'app-login',
  imports: [CommonModule, FormsModule, ReactiveFormsModule, LucideAngularModule],
  templateUrl: './login.html',
  styleUrl: './login.css'
})
export class Login implements OnInit, OnDestroy {
  error: any;
  loginForm!: FormGroup;
  submitted: boolean = false;
  env: any = environment;
  subscriptions: any = [];
  returnTo: string = "/";

  constructor(public api: ApiService, private formBuilder: FormBuilder, public router: Router,
    public route: ActivatedRoute) {
    if (this.api.isAuthenticated()) {
      console.log("user already logged");
      this.router.navigateByUrl("/");
    }
  }

  ngOnInit(): void {
    this.api.creds.clear()
    this.loginForm = this.formBuilder.group({
      username: new FormControl('', Validators.required),
      password: new FormControl('', Validators.required),
      url: [this.api.settings.URL(), Validators.required]
    });

    this.subscriptions = [
      this.api.onLoggedOut.subscribe(error => {
        this.error = error;
        console.log("logged out:" + error);
      }),
      this.api.onLoggedIn.subscribe(() => {
        console.log("logged in");
        this.router.navigate(['/dashboard']);
      })
    ];

  }

  ngOnDestroy() {
    for (let i = 0; i < this.subscriptions.length; i++) {
      this.subscriptions[i].unsubscribe();
    }
    this.subscriptions = [];
  }

  get form() {
    return this.loginForm.controls;
  }
  get username() {
    return this.form['username'] as FormControl;
  }
  get password() {
    return this.form['password'] as FormControl;
  }

  onSubmit() {
    this.error = null;
    this.submitted = true;
    if (!this.loginForm.invalid) {
      try {
        const urlInput = this.form['url'].value.trim();

        if (!/^https?:\/\/[^ "]+$/.test(urlInput)) {
          this.error = "Invalid URL format. Example: http://127.0.0.1:8080/api";
          return;
        }

        const parsed = new URL(urlInput);

        console.log("Parsed URL:", parsed);

        this.api.settings.schema = parsed.protocol;
        this.api.settings.host = parsed.hostname;
        this.api.settings.port = parsed.port;
        this.api.settings.path = parsed.pathname;

        console.log("API Base URL:", this.api.settings.URL());

        this.api
          .login(this.form['username'].value, this.form["password"].value)
          .subscribe({
            next: () => {
              console.log("Login successful");
            },
            error: (err: HttpErrorResponse) => {
              this.error = err;
              console.error("Login failed:", err);
            }
          });

      } catch (err) {
        console.error("URL error:", err);
        this.error = "Invalid URL. Please check the format (e.g., http://127.0.0.1:8080/api).";
      }
    }
  }
}
