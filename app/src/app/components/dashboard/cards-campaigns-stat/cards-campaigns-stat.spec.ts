import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CardsCampaignsStat } from './cards-campaigns-stat';

describe('CardsCampaignsStat', () => {
  let component: CardsCampaignsStat;
  let fixture: ComponentFixture<CardsCampaignsStat>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CardsCampaignsStat]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CardsCampaignsStat);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
