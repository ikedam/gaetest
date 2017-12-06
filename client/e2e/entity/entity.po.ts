import * as moment from 'moment';
import { browser, by, element } from 'protractor';

export class EntityListPage {
  path = '/entity/list';
  addLink = element(by.css('.add-link'));
  entityNames = element.all(by.css('.entity-list tbody .entity-name'));

  navigateTo() {
    return browser.get(this.path);
  }

  isStayIn() {
    return browser.getCurrentUrl().then(
      url => {
        return url.endsWith(this.path);
      }
    );
  }
}

export class EntityNewPage {
  path = '/entity/new';
  nameField = element(by.css('[name=name]'));
  scheduledDateField = element(by.css('[name=scheduledDate]'));
  submitButton = element(by.css('button[type=submit]'));
  validationErrors = element.all(by.css('.form-text.text-danger'));

  navigateTo() {
    return browser.get(this.path);
  }

  isStayIn() {
    return browser.getCurrentUrl().then(
      url => {
        return url.endsWith(this.path);
      }
    );
  }

  setScheduledDate(year: number, month: number, day: number) {
    return this.setScheduledDateByMoment(moment([year, month - 1, day]));
  }

  setScheduledDateByMoment(m: moment.Moment) {
    const dateString = m.format('YYYY-MM-DD');

    // type="date" field doesn't accept clear() and sendKeys(), and the input format depends on locales,
    // modify the type to text and input value.
    // https://github.com/angular/protractor/issues/562
    // https://bugs.chromium.org/p/chromedriver/issues/detail?id=1319
    return browser.executeScript('arguments[0].type = "text";', this.scheduledDateField).then(() => {
      return this.scheduledDateField.clear();
    }).then(() => {
      return this.scheduledDateField.sendKeys(dateString);
    }).then(() => {
      return browser.executeScript('arguments[0].type = "date";', this.scheduledDateField);
    });
  }
}
