const letters = (
  'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  + '!"#$%&\'()=~|{}`*+<>?_-^\\@[;:],./'
);

export function generateRandomString(length: number): string {
  let ret = '';
  for (let i = 0; i < length; ++i) {
    ret += letters.charAt(Math.floor(Math.random() * letters.length));
  }
  return ret;
}
