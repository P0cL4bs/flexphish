// biome-ignore lint/style/useImportType: <explanation>
import { HttpClient, HttpErrorResponse, HttpHeaders, HttpParams } from '@angular/common/http';
import { inject } from '@angular/core'
import { EventEmitter, Injectable } from '@angular/core';
import { EMPTY, concat, lastValueFrom, of, throwError } from 'rxjs'
import { catchError, map, tap } from 'rxjs/operators';
import { from } from 'rxjs';
import { Observable } from 'rxjs';
import { interval } from "rxjs/internal/observable/interval";
import { startWith, switchMap } from "rxjs/operators";
import { TokenResponse } from '../models/token';
import { jwtDecode } from 'jwt-decode';
import { Template, TemplateCreateRequest, TemplateDeleteRequest, TemplateHtmlFile, TemplateHtmlFileDeleteRequest, TemplateHtmlFileUpdateRequest, TemplateHtmlFileUploadRequest, TemplateMetadata, TemplateStaticFile, TemplateStaticFileRequest, TemplateUpdateRequest } from '../models/template.model';
import { TemplatesResponse } from '../models/template-response.model';
import { CampaignDetail } from '../models/campaign-detail.model';
import { Campaign, CreateCampaignRequest, UpdateCampaignRequest } from '../models/campaign.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { Config } from '../models/config.model';
import { CampaignAnalytics } from '../models/campaign-analytics.model';
import { CreateGroupRequest, Group, GroupTarget, GroupTargetPayload, UpdateGroupRequest } from '../models/group.model';
import { SMTPProfile, SMTPProfilePayload, SMTPTestPayload } from '../models/smtp.model';
import { EmailTemplate, EmailTemplateAttachment, EmailTemplatePayload } from '../models/email-template.model';


interface JWTPayload {
    exp?: number;
    [key: string]: any;
}



export class Settings {
    public schema: string = 'http:';
    public host: string = "127.0.0.1";
    public port: string = '8088';
    public path = '/api';

    constructor() {
        const stored = localStorage.getItem('settings');
        if (stored) {
            try {
                const obj = JSON.parse(stored);
                this.from(obj);
            } catch (e) {
                console.warn("Settings invalid in storage:", e);
            }
        }
    }

    public URL(): string {
        return `${this.schema}//${this.host}:${this.port}${this.path}`;
    }

    public Warning(): boolean {
        if (this.host === 'localhost' || this.host === '127.0.0.1')
            return false;

        return this.schema !== 'https:';
    }

    public from(obj: any) {
        this.schema = obj.schema ?? this.schema;
        this.host = obj.host ?? this.host;
        this.port = obj.port ?? this.port;
        this.path = obj.path ?? this.path;
    }

    public save() {
        localStorage.setItem('settings', JSON.stringify({
            schema: this.schema,
            host: this.host,
            port: this.port,
            path: this.path,
        }));
    }
}

export class Credentials {
    public valid = false;
    public token: string = "";
    public headers: HttpHeaders = new HttpHeaders();

    public setToken(token: string) {
        this.token = token;
        this.headers = new HttpHeaders().set("Authorization", `Bearer ${this.token}`);
    }

    public isValidToken(): boolean {
        const token = localStorage.getItem('auth_token');
        if (!token) return false;

        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            const exp = payload.exp * 1000;

            return Date.now() < exp;
        } catch {
            return false;
        }
    }

    public load() {
        const saved = localStorage.getItem('auth_token');
        if (saved) {
            this.setToken(saved);
        }
    }

    public getToken(): string | null {
        return localStorage.getItem('auth_token');
    }

    public save() {
        localStorage.setItem('auth_token', this.token);
    }

    public clear() {
        this.token = "";
        this.valid = false;
        this.headers = new HttpHeaders();
        localStorage.removeItem('auth_token');
    }
}

@Injectable({
    providedIn: 'root'
})
export class ApiService {
    public settings: Settings = new Settings();
    public creds: Credentials = new Credentials();
    public events: Event[] = new Array();
    public error: any = null;

    public onLoggedOut: EventEmitter<any> = new EventEmitter();
    public onLoggedIn: EventEmitter<any> = new EventEmitter();

    constructor(private http: HttpClient) {
        this.creds.load();
    }

    public isAuthenticated(): boolean {
        return this.creds.valid;
    }

    public isValidJwtToken(): boolean {
        var token = this.creds.getToken()
        if (!token) return false;

        try {
            const payload: JWTPayload = jwtDecode(token);
            return !!payload.exp;
        } catch (err) {
            console.warn("Invalid JWT format", err);
            return false;
        }
    }
    public login(email: string, password: string) {
        return this.http
            .post<TokenResponse>(`${this.settings.URL()}/login`, { email, password })
            .pipe(
                tap(response => {
                    this.creds.setToken(response.token);
                    this.creds.save();
                    this.setLoggedIn();
                }),
                map(() => true),
                catchError((error: HttpErrorResponse) => {
                    console.error("Login error:", error);
                    return throwError(() => error);
                })
            );
    }
    public validateToken() {
        return this.http.get<{ valid: boolean }>(`${this.settings.URL()}/auth/validate`, { headers: this.creds.headers }).pipe(
            map(response => response.valid === true),
            catchError((error: HttpErrorResponse) => {
                console.warn("Token validation failed:", error);
                return of(false);
            })
        );
    }

    public getTemplates(): Observable<TemplatesResponse> {
        return this.http.get<TemplatesResponse>(
            `${this.settings.URL()}/templates`
            , { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getTemplateById(id: string): Observable<Template> {
        return this.http.get<Template>(
            `${this.settings.URL()}/templates/${id}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }
    public getTemplateMetadataById(template_id: string): Observable<TemplateMetadata> {
        return this.http.get<TemplateMetadata>(
            `${this.settings.URL()}/templates/${template_id}/metadata`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteTemplate(data: TemplateDeleteRequest): Observable<any> {
        return this.http.delete<any>(
            `${this.settings.URL()}/templates`,
            {
                headers: this.creds.headers,
                body: data
            }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getTemplatesList(): Observable<TemplateMetadata[]> {
        return this.getTemplates().pipe(
            map(response => Object.values(response.templates))
        );
    }

    public updateTemplate(
        data: TemplateUpdateRequest
    ): Observable<any> {

        return this.http.put(
            `${this.settings.URL()}/templates`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }
    public createTemplate(
        data: TemplateCreateRequest
    ): Observable<any> {

        return this.http.post(
            `${this.settings.URL()}/templates`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public getTemplateHtmlFiles(template_id: string): Observable<TemplateHtmlFile[]> {

        return this.http.get<TemplateHtmlFile[]>(
            `${this.settings.URL()}/templates/${template_id}/html-files`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public updateTemplateHtmlFile(data: TemplateHtmlFileUpdateRequest, template_id: string): Observable<any> {

        return this.http.put(
            `${this.settings.URL()}/templates/${template_id}/html-files`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public uploadTemplateHtmlFile(
        data: TemplateHtmlFileUploadRequest,
        template_id: string
    ): Observable<any> {

        return this.http.post(
            `${this.settings.URL()}/templates/${template_id}/html-files`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteTemplateHtmlFile(
        data: TemplateHtmlFileDeleteRequest,
        template_id: string
    ): Observable<any> {

        return this.http.request(
            'DELETE',
            `${this.settings.URL()}/templates/${template_id}/html-files`,
            {
                body: data,
                headers: this.creds.headers
            }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public getTemplateStaticFiles(template_id: string): Observable<TemplateStaticFile[]> {

        return this.http.get<TemplateStaticFile[]>(
            `${this.settings.URL()}/templates/${template_id}/static-files`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }
    public createTemplateStaticFile(data: TemplateStaticFileRequest): Observable<any> {

        return this.http.post(
            `${this.settings.URL()}/templates/${data.t_filename}/static-files`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public updateTemplateStaticFile(data: TemplateStaticFileRequest): Observable<any> {

        return this.http.put(
            `${this.settings.URL()}/templates/${data.t_filename}/static-files`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public deleteTemplateStaticFile(data: TemplateStaticFileRequest): Observable<any> {

        return this.http.request(
            'DELETE',
            `${this.settings.URL()}/templates/${data.t_filename}/static-files`,
            {
                body: data,
                headers: this.creds.headers
            }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    getCampaigns(
        page: number = 1,
        pageSize: number = 10
    ): Observable<PaginatedResponse<Campaign>> {
        return this.http.get<PaginatedResponse<Campaign>>(
            `${this.settings.URL()}/campaigns`,

            {
                headers: this.creds.headers,
                params: {
                    page,
                    page_size: pageSize
                }
            }
        );
    }

    getCampaignAnalytics(
        period: string = 'day'
    ): Observable<CampaignAnalytics> {

        return this.http.get<CampaignAnalytics>(
            `${this.settings.URL()}/campaigns/analytics`,
            {
                headers: this.creds.headers,
                params: {
                    period
                }
            }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    /**
     * GET /api/campaigns/:id
     */
    getCampaignById(id: number): Observable<CampaignDetail> {
        return this.http.get<CampaignDetail>(
            `${this.settings.URL()}/campaigns/${id}`
            , { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    updateCampaign(
        id: number,
        payload: UpdateCampaignRequest
    ): Observable<CampaignDetail> {

        return this.http.put<CampaignDetail>(
            `${this.settings.URL()}/campaigns/${id}`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public createCampaign(
        payload: CreateCampaignRequest
    ): Observable<Campaign> {

        return this.http.post<Campaign>(
            `${this.settings.URL()}/campaigns`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    deleteCampaign(id: number): Observable<void> {

        return this.http.delete<void>(
            `${this.settings.URL()}/campaigns/${id}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }
    deleteResult(campaignId: number, resultId: number): Observable<void> {

        return this.http.delete<void>(
            `${this.settings.URL()}/campaigns/${campaignId}/results/${resultId}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public startCampaign(id: number): Observable<CampaignDetail> {

        return this.http.post<CampaignDetail>(
            `${this.settings.URL()}/campaigns/${id}/start`,
            {},
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }


    public stopCampaign(id: number): Observable<CampaignDetail> {

        return this.http.post<CampaignDetail>(
            `${this.settings.URL()}/campaigns/${id}/stop`,
            {},
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }


    public completeCampaign(id: number): Observable<CampaignDetail> {

        return this.http.post<CampaignDetail>(
            `${this.settings.URL()}/campaigns/${id}/complete`,
            {},
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public logout() {
        console.log("api.logout()");
        this.clearStorage();
        this.creds.valid = false;
        this.creds.clear()
    }

    private clearStorage() {
        console.log("api.clearStorage()");

        localStorage.removeItem('auth');

        this.creds.clear();
    }

    private saveStorage() {
        this.creds.save();
        this.settings.save();
    }


    private setLoggedIn(): boolean {
        const wasLogged = this.creds.valid;

        this.creds.valid = true;
        this.saveStorage();

        if (wasLogged === false) {
            console.log("setLoggedIn: emit");
            this.onLoggedIn.emit();
        }

        return wasLogged;
    }
    public getConfigs(): Observable<Config> {

        return this.http.get<Config>(
            `${this.settings.URL()}/configs`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public updateConfigs(data: Record<string, any>): Observable<any> {

        return this.http.put<any>(
            `${this.settings.URL()}/configs`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );

    }

    public getGroups(): Observable<Group[]> {
        return this.http.get<Group[]>(
            `${this.settings.URL()}/groups`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public createGroup(payload: CreateGroupRequest): Observable<Group> {
        return this.http.post<Group>(
            `${this.settings.URL()}/groups`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public updateGroup(id: number, payload: UpdateGroupRequest): Observable<Group> {
        return this.http.put<Group>(
            `${this.settings.URL()}/groups/${id}`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteGroup(id: number): Observable<void> {
        return this.http.delete<void>(
            `${this.settings.URL()}/groups/${id}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getGroupTargets(groupId: number): Observable<GroupTarget[]> {
        return this.http.get<GroupTarget[]>(
            `${this.settings.URL()}/groups/${groupId}/targets`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public createGroupTarget(groupId: number, payload: GroupTargetPayload): Observable<GroupTarget> {
        return this.http.post<GroupTarget>(
            `${this.settings.URL()}/groups/${groupId}/targets`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public updateGroupTarget(groupId: number, targetId: number, payload: GroupTargetPayload): Observable<GroupTarget> {
        return this.http.put<GroupTarget>(
            `${this.settings.URL()}/groups/${groupId}/targets/${targetId}`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteGroupTarget(groupId: number, targetId: number): Observable<void> {
        return this.http.delete<void>(
            `${this.settings.URL()}/groups/${groupId}/targets/${targetId}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getSMTPProfiles(): Observable<SMTPProfile[]> {
        return this.http.get<SMTPProfile[]>(
            `${this.settings.URL()}/smtp-profiles`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public createSMTPProfile(payload: SMTPProfilePayload): Observable<SMTPProfile> {
        return this.http.post<SMTPProfile>(
            `${this.settings.URL()}/smtp-profiles`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public updateSMTPProfile(id: number, payload: SMTPProfilePayload): Observable<SMTPProfile> {
        return this.http.put<SMTPProfile>(
            `${this.settings.URL()}/smtp-profiles/${id}`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteSMTPProfile(id: number): Observable<void> {
        return this.http.delete<void>(
            `${this.settings.URL()}/smtp-profiles/${id}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public testSMTPProfile(payload: SMTPTestPayload): Observable<{ message: string }> {
        return this.http.post<{ message: string }>(
            `${this.settings.URL()}/smtp-profiles/test`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getEmailTemplates(): Observable<EmailTemplate[]> {
        return this.http.get<EmailTemplate[]>(
            `${this.settings.URL()}/email-templates`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public createEmailTemplate(payload: EmailTemplatePayload): Observable<EmailTemplate> {
        return this.http.post<EmailTemplate>(
            `${this.settings.URL()}/email-templates`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public updateEmailTemplate(id: number, payload: EmailTemplatePayload): Observable<EmailTemplate> {
        return this.http.put<EmailTemplate>(
            `${this.settings.URL()}/email-templates/${id}`,
            payload,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteEmailTemplate(id: number): Observable<void> {
        return this.http.delete<void>(
            `${this.settings.URL()}/email-templates/${id}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public getEmailTemplateAttachments(templateId: number): Observable<EmailTemplateAttachment[]> {
        return this.http.get<EmailTemplateAttachment[]>(
            `${this.settings.URL()}/email-templates/${templateId}/attachments`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public uploadEmailTemplateAttachment(templateId: number, file: File): Observable<EmailTemplateAttachment> {
        const data = new FormData();
        data.append('file', file);

        return this.http.post<EmailTemplateAttachment>(
            `${this.settings.URL()}/email-templates/${templateId}/attachments`,
            data,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

    public deleteEmailTemplateAttachment(templateId: number, attachmentId: number): Observable<void> {
        return this.http.delete<void>(
            `${this.settings.URL()}/email-templates/${templateId}/attachments/${attachmentId}`,
            { headers: this.creds.headers }
        ).pipe(
            catchError((error: HttpErrorResponse) => {
                return throwError(() => error);
            })
        );
    }

}
