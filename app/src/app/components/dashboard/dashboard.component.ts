import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router, RouterOutlet } from '@angular/router';
import { LucideAngularModule } from 'lucide-angular';
import { ApiService } from 'src/app/services/api.service';
import { BaseChartDirective } from 'ng2-charts';
import { ColumnSortedEvent, SortService } from 'src/app/services/sort.service';
import { FormsModule } from '@angular/forms';
import { CampaignEventsChart } from "./campaign-events-chart/campaign-events-chart";
import { EventsTimelineChart } from "./events-timeline-chart/events-timeline-chart";
import { EventTypesChart } from "./event-types-chart/event-types-chart";
import { TopCampaignsChart } from "./top-campaigns-chart/top-campaigns-chart";
import { CardsCampaignsStat } from "./cards-campaigns-stat/cards-campaigns-stat";



@Component({
  standalone: true,
  selector: 'app-dashboard',
  imports: [LucideAngularModule, CommonModule, FormsModule, CampaignEventsChart, EventsTimelineChart, EventTypesChart, TopCampaignsChart, CardsCampaignsStat],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent implements OnInit, OnDestroy {
  uptime: string = '';
  @ViewChild(BaseChartDirective) chart?: BaseChartDirective;

  constructor(public api: ApiService, private sortService: SortService, public router: Router,
    public route: ActivatedRoute) {
  }

  ngOnInit() {
  }
  ngOnDestroy() {
  }

}
