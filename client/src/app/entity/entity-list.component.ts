import { Component, OnInit } from '@angular/core';

import { NgbModule } from '@ng-bootstrap/ng-bootstrap';

import { Entity } from './entity';

@Component({
  templateUrl: './entity-list.component.html',
  styleUrls: ['./entity-list.component.css']
})
export class EntityListComponent implements OnInit {
  entityList: Entity[] = [];

  ngOnInit() {
    {
      const entity = new Entity();
      entity.id = 1;
      entity.name = 'テスト1';
      entity.createdAt = new Date();
      this.entityList.push(entity);
    }
    {
      const entity = new Entity();
      entity.id = 2;
      entity.name = 'テスト2';
      entity.createdAt = new Date();
      this.entityList.push(entity);
    }
    {
      const entity = new Entity();
      entity.id = 3;
      entity.name = 'テスト3';
      entity.createdAt = new Date();
      this.entityList.push(entity);
    }
  }
}
