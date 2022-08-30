const path = require('path')
const os = require('os')

module.exports = {
  EXPIRES_IN: 3600,
  CONFIG_DIR: path.join(os.homedir(), '.config/gyuma'),
  TOKEN_FILE: 'token.json',
  CREDENTIALS_FILE: 'credentials',
}
