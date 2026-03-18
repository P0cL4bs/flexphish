import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CampaignView } from './campaign-view';

describe('CampaignView', () => {
  let component: CampaignView;
  let fixture: ComponentFixture<CampaignView>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CampaignView]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CampaignView);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
