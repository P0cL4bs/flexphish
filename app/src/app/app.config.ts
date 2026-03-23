import { type ApplicationConfig, importProvidersFrom } from '@angular/core'
import { provideRouter } from '@angular/router'

import { provideHttpClient, withInterceptors } from '@angular/common/http'
import { ArchiveRestore, ArrowUpRight, Binoculars, ChevronRight, CircleCheck, CirclePlay, CirclePlus, CircleStop, CircleX, Component, Copy, Cpu, Crosshair, File, FileCode, FileJson2, FilePlus, FileSpreadsheet, Gauge, Globe, Home, LaptopMinimalCheck, LogIn, LogOut, Logs, LucideAngularModule, MemoryStick, Menu, MonitorCog, MonitorSmartphone, Network, Octagon, Package, PanelTopClose, PanelTopOpen, Pause, Play, Radar, RadioTower, RotateCcw, RouteOff, SatelliteDish, Save, ServerCrash, Settings, Signal, SignalHigh, SignalLow, SignalMedium, SquareActivity, SquareAsterisk, SquareChevronLeft, SquareTerminal, Squircle, Terminal, Trash, Trash2, UserCheck, Wifi, WifiOff, Antenna, Fan, SquareChevronRight, Megaphone, LayoutTemplate, ChartNoAxesCombined, Key, Flame, ChevronDown, ShieldCheck, Activity, MousePointerClick, KeyRound, Layers, Search, Folder, Users, SquarePen, Server, Mailbox, EllipsisVertical, CalendarClock } from "lucide-angular";
import { routes } from './app.routes'
import { provideHighlightOptions } from 'ngx-highlightjs';
import { provideCharts, withDefaultRegisterables } from 'ng2-charts';
import { authInterceptor } from './core/interceptors/auth-interceptor';

export const appConfig: ApplicationConfig = {
  providers: [provideRouter(routes),
  provideHttpClient(
    withInterceptors([authInterceptor])
  ),
  importProvidersFrom(LucideAngularModule.pick({ File, Home, Menu, UserCheck, ArrowUpRight, Logs, SquareTerminal, Terminal, Play, Squircle, LogOut, SquareChevronLeft, SquareChevronRight, Gauge, Settings, Component, ServerCrash, Package, Save, PanelTopClose, FilePlus, Trash, FileCode, MonitorSmartphone, SquareActivity, ArchiveRestore, Trash2, ChevronRight, Wifi, Network, PanelTopOpen, LogIn, SignalHigh, SignalLow, SignalMedium, Signal, Radar, RouteOff, Crosshair, RadioTower, WifiOff, CircleX, CircleCheck, Globe, SatelliteDish, Copy, Binoculars, CirclePlus, SquareAsterisk, FileJson2, FileSpreadsheet, LaptopMinimalCheck, Cpu, MemoryStick, MonitorCog, RotateCcw, Antenna, Fan, Megaphone, LayoutTemplate, ChartNoAxesCombined, Key, Flame, ChevronDown, ShieldCheck, Activity, MousePointerClick, KeyRound, Layers, Search, Folder, Users, SquarePen, Server, Mailbox, EllipsisVertical, CalendarClock })),

  provideHighlightOptions({
    fullLibraryLoader: () => import('highlight.js'),
    themePath: 'assets/styles/dark.css',
  }),
  provideCharts(withDefaultRegisterables()),
  ],

}
