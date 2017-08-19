import { Injectable } from '@angular/core';
import { Headers, Http } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/toPromise';

import { environment } from '../../environments/environment';
import { Entity } from './entity';

@Injectable()
export class EntityService {
  constructor (private http: Http) {}

  getEntityList(): Promise<Entity[]> {
    return this.http.get(`${environment.apiURL}/entity/`).toPromise().then(
      response => {
        let data = response.json();
        return data.map(d => {return new Entity(d);});
      }
    );
  }

  createEntity(entity: Entity): Promise<Entity> {
    return this.http.post(`${environment.apiURL}/entity/`, entity, {
      headers: new Headers({'Content-Type': 'application/json'}),
    }).toPromise().then(
      response => {
        let data = response.json();
        return new Entity(data);
      }
    );
  }
}
