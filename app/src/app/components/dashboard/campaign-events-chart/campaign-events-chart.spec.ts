import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CampaignEventsChart } from './campaign-events-chart';

describe('CampaignEventsChart', () => {
  let component: CampaignEventsChart;
  let fixture: ComponentFixture<CampaignEventsChart>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CampaignEventsChart]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CampaignEventsChart);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
