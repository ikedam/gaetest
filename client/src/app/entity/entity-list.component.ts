import { Component, OnInit } from '@angular/core';

import { NgbModule } from '@ng-bootstrap/ng-bootstrap';

import { Entity } from './entity';
import { EntityService } from './entity.service';

@Component({
  templateUrl: './entity-list.component.html',
  styleUrls: ['./entity-list.component.css']
})
export class EntityListComponent implements OnInit {
  entityList: Entity[] = [];

  constructor(private entityService: EntityService) {
  }

  ngOnInit() {
    this.entityService.getEntityList().then(
      entityList => {
        this.entityList = entityList;
      }
    );
    /*
    this.entityList = [
      new Entity({name: 'テスト1',}),
      new Entity({name: 'テスト2',}),
      new Entity({name: 'テスト3',}),
    ];
    */
  }
}
