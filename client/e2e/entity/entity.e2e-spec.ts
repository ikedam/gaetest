import * as moment from 'moment';
import { browser, protractor } from 'protractor';
import { EntityListPage, EntityNewPage } from './entity.po';
import { generateRandomString } from '../test-util';

describe('Entity', () => {
  it('Can move to entity page', () => {
    const listPage = new EntityListPage();
    listPage.navigateTo();
    listPage.addLink.click();

    const newPage = new EntityNewPage();
    expect(newPage.isStayIn()).toEqual(true);
  });
  it('Can add a new entity', () => {
    const newPage = new EntityNewPage();
    newPage.navigateTo();
    const name = generateRandomString(10);
    newPage.nameField.sendKeys(name);
    newPage.setScheduledDateByMoment(moment().add(1, 'd'));
    newPage.submitButton.click();

    const listPage = new EntityListPage();
    expect(listPage.isStayIn()).toEqual(true);
    expect(listPage.entityNames.get(0).getText()).toEqual(name);
  });
  it('Form validation', () => {
    const newPage = new EntityNewPage();
    newPage.navigateTo();
    expect(newPage.submitButton.isEnabled()).toEqual(false);
    expect(newPage.validationErrors.isPresent()).toEqual(false);

    newPage.nameField.sendKeys('test');
    newPage.setScheduledDateByMoment(moment().add(1, 'd'));

    expect(newPage.submitButton.isEnabled()).toEqual(true);
    expect(newPage.validationErrors.isPresent()).toEqual(false);

    // newPage.nameField.clear() looks not update the model
    newPage.nameField.sendKeys(
      protractor.Key.BACK_SPACE
      + protractor.Key.BACK_SPACE
      + protractor.Key.BACK_SPACE
      + protractor.Key.BACK_SPACE
    );

    expect(newPage.submitButton.isEnabled()).toEqual(false);

    newPage.nameField.sendKeys('test');
    expect(newPage.submitButton.isEnabled()).toEqual(true);

    newPage.setScheduledDateByMoment(moment().add(-1, 'd'));
    expect(newPage.submitButton.isEnabled()).toEqual(false);

    newPage.setScheduledDateByMoment(moment().add(1, 'd'));
    expect(newPage.submitButton.isEnabled()).toEqual(true);

    // There looks no appropriate way to set touched
    // expect(newPage.validationErrors.isPresent()).toEqual(true);
  });
});
