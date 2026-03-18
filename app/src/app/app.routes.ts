import type { Routes } from '@angular/router'
import { Login } from './components/login/login';
import { AuthGuard } from './auth.guard';
import { DashboardComponent } from './components/dashboard/dashboard.component';
import { TemplatesView } from './components/templates-view/templates-view';
import { CampaignView } from './components/campaign-view/campaign-view';
import { CampaignDetailView } from './components/campaign-detail-view/campaign-detail-view';
import { TemplateViewByID } from './components/template-view-by-id/template-view-by-id';
import { ConfigView } from './components/config-view/config-view';

export const routes: Routes = [
  {
    path: 'login',
    component: Login,
  },
  {
    path: 'dashboard',
    component: DashboardComponent, canActivate: [AuthGuard]
  },
  {
    path: 'templates',
    component: TemplatesView, canActivate: [AuthGuard],
    children: [
      {
        path: ':filename',
        component: TemplateViewByID, canActivate: [AuthGuard]
      }
    ]
  },
  {
    path: 'campaigns',
    component: CampaignView, canActivate: [AuthGuard]
  },
  {
    path: 'campaigns/:id',
    component: CampaignDetailView, canActivate: [AuthGuard]
  },
  {
    path: 'config',
    component: ConfigView, canActivate: [AuthGuard]
  },
  { path: '**', redirectTo: 'dashboard' }
];
