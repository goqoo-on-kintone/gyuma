const path = require('path')
const fs = require('fs')
const { CONFIG_DIR, TOKEN_FILE, CREDENTIALS_FILE } = require('./constants')
const { decrypt } = require('./encrypt')

const decryptCredentias = (encryptedCredentials, password) => {
  const decreptedStr = decrypt(encryptedCredentials, password)
  const decreptedObj = JSON.parse(decreptedStr)
  return decreptedObj
}

module.exports = ({ domain, password }) => {
  // TODO: 最優先：write-config/read-configモジュールを統合

  // TODO: パスワードが指定されていなければ標準入力させる
  // TODO: tokenとcredentialsで関数を分けてそれぞれexports

  // 保存済みトークン
  const tokenFilePath = path.join(CONFIG_DIR, domain, TOKEN_FILE)
  const token = fs.existsSync(tokenFilePath) ? require(tokenFilePath) : {}

  // 保存済みクレデンシャル
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  const encryptedCredentials = fs.existsSync(credentialsFilePath) && fs.readFileSync(credentialsFilePath, 'utf-8')
  const credentials = encryptedCredentials ? decryptCredentias(encryptedCredentials, password) : {}

  return { token, credentials }
}
