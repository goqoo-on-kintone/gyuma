const path = require('path')
const fs = require('fs')
const { CONFIG_DIR, TOKEN_FILE, CREDENTIALS_FILE } = require('./constants')

module.exports = domain => {
  // 保存済みトークン
  const tokenFilePath = path.join(CONFIG_DIR, domain, TOKEN_FILE)
  const token = fs.existsSync(tokenFilePath) ? require(tokenFilePath) : {}

  // 保存済みクレデンシャル情報(client_id, etc...)
  // TODO: credentialsを複合化
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  const credentials = fs.existsSync(credentialsFilePath) ? require(credentialsFilePath) : {}

  return { token, credentials }
}
