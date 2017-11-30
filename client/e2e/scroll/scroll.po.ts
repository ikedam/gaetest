import { browser, by, element } from 'protractor';
import { promise as wdpromise } from 'selenium-webdriver';

class TextRectanble {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

export class ScrollPage {
  path = '/scroll';
  generateButton = element(by.css('[data-test-label=generate]'));
  scroller = element(by.css('.scroller'));
  activePoint = element(by.css('.active-point'));

  navigateTo() {
    return browser.get(this.path);
  }

  isActivePointVisible(): wdpromise.Promise<boolean> {
    return browser.executeScript('return document.querySelector(".scroller").getBoundingClientRect();').then(
      (scrollerRect: TextRectanble) => {
        return browser.executeScript('return document.querySelector(".active-point").getBoundingClientRect();').then(
          (activePointRect: TextRectanble) => {
            return (
              scrollerRect.top <= activePointRect.top
              && activePointRect.bottom <= scrollerRect.bottom
            );
          }
       );
    });
  }
}
