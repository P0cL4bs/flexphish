import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CampaignDetailView } from './campaign-detail-view';

describe('CampaignDetailView', () => {
  let component: CampaignDetailView;
  let fixture: ComponentFixture<CampaignDetailView>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CampaignDetailView]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CampaignDetailView);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
