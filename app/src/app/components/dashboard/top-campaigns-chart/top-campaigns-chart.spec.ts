import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TopCampaignsChart } from './top-campaigns-chart';

describe('TopCampaignsChart', () => {
  let component: TopCampaignsChart;
  let fixture: ComponentFixture<TopCampaignsChart>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopCampaignsChart]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TopCampaignsChart);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
