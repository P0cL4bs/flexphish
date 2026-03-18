import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EventTypesChart } from './event-types-chart';

describe('EventTypesChart', () => {
  let component: EventTypesChart;
  let fixture: ComponentFixture<EventTypesChart>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventTypesChart]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EventTypesChart);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
