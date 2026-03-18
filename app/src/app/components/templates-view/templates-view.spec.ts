import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TemplatesView } from './templates-view';

describe('TemplatesView', () => {
  let component: TemplatesView;
  let fixture: ComponentFixture<TemplatesView>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TemplatesView]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TemplatesView);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
