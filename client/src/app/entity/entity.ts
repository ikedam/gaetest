import * as moment from 'moment'

export class Entity {
  id: number;
  name: string;
  createdAt: Date;

  constructor(json: {[key: string]: any} = null) {
    this.createdAt = new Date();
    if (json) {
      this.id = json['id'];
      this.name = json['name'];
      if (json['createdAt']) {
        if (json['createdAt'] instanceof Date) {
          this.createdAt = json['createdAt'];
        } else {
          this.createdAt = moment(json['createdAt']).toDate();
        }
      }
    }
  }
}
