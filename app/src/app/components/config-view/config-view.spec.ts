import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ConfigView } from './config-view';

describe('ConfigView', () => {
  let component: ConfigView;
  let fixture: ComponentFixture<ConfigView>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ConfigView]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ConfigView);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
