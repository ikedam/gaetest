import { Component } from '@angular/core';

import { Entity } from './entity';

@Component({
  templateUrl: './entity-list.component.html',
  styleUrls: ['./entity-list.component.css']
})
export class EntityListComponent {
  entityList: Entity[] = [];
}
