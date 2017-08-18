import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

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

  constructor(private builder: FormBuilder) {
  }

  ngOnInit() {
    this.form = this.builder.group(this.formDef);
  }

  submit() {
    this.submitted = true;
  }
}
