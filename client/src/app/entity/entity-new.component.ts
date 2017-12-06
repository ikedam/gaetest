import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import * as moment from 'moment';

import { Entity } from './entity';
import { EntityService } from './entity.service';

@Component({
  templateUrl: './entity-new.component.html',
  styleUrls: ['./entity-new.component.css']
})
export class EntityNewComponent implements OnInit {
  formDef = {
    name: ['', [Validators.required, ], ],
    scheduledDate: ['', [Validators.required, EntityNewComponent.validateDateAfterNow, ], ],
  };

  form: FormGroup;

  submitted = false;

  constructor(
    private builder: FormBuilder,
    private router: Router,
    private activatedRoute: ActivatedRoute,
    private entityService: EntityService,
  ) {
  }

  static validateDateAfterNow(control: AbstractControl): {[key: string]: any} {
    return EntityNewComponent.validateDateAfterNowForValue(control.value);
  }

  static validateDateAfterNowForValue(value: string): {[key: string]: any} {
    const m = moment(value);
    if (!m.isValid()) {
      // invalid value
      return null;
    }
    const inputDate = m.toDate();
    const now = new Date();
    if (inputDate.getTime() < now.getTime()) {
      return {
        'dateAfterNow': {
          'value': value,
        },
      };
    }
    return null;
  }

  ngOnInit() {
    this.form = this.builder.group(this.formDef);
  }

  submit() {
    this.submitted = true;
    this.entityService.createEntity(new Entity(this.form.value)).then(
      () => {
        this.router.navigate(['../list'], {relativeTo: this.activatedRoute});
      }
    );
  }
}
