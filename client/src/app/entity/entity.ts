import * as moment from 'moment';

export class Entity {
  id: number;
  name: string;
  scheduledDate: Date;
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
      if (json['scheduledDate']) {
        if (json['scheduledDate'] instanceof Date) {
          this.scheduledDate = json['scheduledDate'];
        } else {
          this.scheduledDate = moment(json['scheduledDate']).toDate();
        }
      }
    }
  }
}
