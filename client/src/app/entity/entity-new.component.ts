import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

import { Entity } from './entity';
import { EntityService } from './entity.service';

@Component({
  templateUrl: './entity-new.component.html',
  styleUrls: ['./entity-new.component.css']
})
export class EntityNewComponent implements OnInit {
  formDef = {
    name: ['', Validators.required, ],
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
