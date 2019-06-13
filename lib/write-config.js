const fs = require('fs')
const path = require('path')
const { CONFIG_DIR, CONFIG_FILE } = require('./constants')

module.exports = token => {
  fs.mkdirSync(CONFIG_DIR, { recursive: true })
  fs.writeFileSync(path.join(CONFIG_DIR, CONFIG_FILE), JSON.stringify(token))
}
