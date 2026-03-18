import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TemplateViewById } from './template-view-by-id';

describe('TemplateViewById', () => {
  let component: TemplateViewById;
  let fixture: ComponentFixture<TemplateViewById>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TemplateViewById]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TemplateViewById);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
