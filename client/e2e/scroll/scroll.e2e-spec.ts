import { browser, protractor } from 'protractor';
import { ScrollPage } from './scroll.po';

describe('Scroll', () => {
  it('Active point is always visible', () => {
    const scrollPage = new ScrollPage();
    scrollPage.navigateTo();
    for (let i = 0; i < 10; ++i) {
      scrollPage.generateButton.click();
      expect(scrollPage.isActivePointVisible()).toEqual(true);
    }
  });
});
