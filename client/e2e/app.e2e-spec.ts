import { AppPage } from './app.po';

describe('client App', () => {
  let page: AppPage;

  beforeEach(() => {
    page = new AppPage();
  });

  it('should display navigation bar', () => {
    page.navigateTo();
    expect(page.getAppTitle()).toEqual('gaetest');
  });
});
