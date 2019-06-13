const fs = require('fs')
const path = require('path')
const { CONFIG_DIR, CONFIG_FILE } = require('./constants')

module.exports = token => {
  // TODO: client_id, client_secretだけはcredentials.jsonに分けて保存＆暗号化
  fs.mkdirSync(CONFIG_DIR, { recursive: true })
  fs.writeFileSync(path.join(CONFIG_DIR, CONFIG_FILE), JSON.stringify(token))
}
