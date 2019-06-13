const path = require('path')
const os = require('os')

module.exports = {
  EXPIRES_IN: 3600,
  CONFIG_DIR: path.join(os.homedir(), '.config/gyuma'),
  CONFIG_FILE: 'gyuma.json',
}
