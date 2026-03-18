import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EventsTimelineChart } from './events-timeline-chart';

describe('EventsTimelineChart', () => {
  let component: EventsTimelineChart;
  let fixture: ComponentFixture<EventsTimelineChart>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EventsTimelineChart]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EventsTimelineChart);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
