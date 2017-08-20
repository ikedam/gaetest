var baseConfig = require('./karma.conf.js');

module.exports = function (config) {
  baseConfig(config);
  config.set({
    customLaunchers: {
      ChromeInDocker: {
        base: 'ChromeHeadless',
        flags: [
          // To avoid the error:
          // Failed to move to new namespace: PID namespaces supported,
          // Network namespace supported, but failed: errno = Operation not permitted
          '--no-sandbox'
        ]
      }
    },
    browsers: ['ChromeInDocker']
  });
};
