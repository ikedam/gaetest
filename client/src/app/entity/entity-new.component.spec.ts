import * as moment from 'moment';

import { EntityNewComponent } from './entity-new.component';

describe('EntityNewComponent', () => {
  it('validateDateAfterNowForValue', () => {
    expect(EntityNewComponent.validateDateAfterNowForValue('')).toEqual(null);
    expect(EntityNewComponent.validateDateAfterNowForValue('2017-01-01')).not.toEqual(null);
    expect(EntityNewComponent.validateDateAfterNowForValue(moment().add(7, 'd').format('YYYY-MM-DD'))).toEqual(null);
  });
});
