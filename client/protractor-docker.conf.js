var baseConfig = require('./protractor.conf.js');

exports.config = {}
for(var key in baseConfig.config) {
  exports.config[key] = baseConfig.config[key];
};

exports.config['capabilities'] =  {
  browserName: 'chrome',
  binary: '/usr/bin/chromium-browser',
  chromeOptions: {
    args: [
      '--headless', '--disable-gpu',
      // To avoid the error:
      // Failed to move to new namespace: PID namespaces supported,
      // Network namespace supported, but failed: errno = Operation not permitted
      '--no-sandbox'
    ]
  }
};
exports.config['chromeDriver'] = '/usr/bin/chromedriver';
