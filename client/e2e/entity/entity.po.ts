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
}
