const fs = require('fs')
const path = require('path')
const moment = require('moment')
const { CONFIG_DIR, TOKEN_FILE, CREDENTIALS_FILE } = require('./constants')
const { encrypt } = require('./encrypt')
const { decrypt } = require('./encrypt')

const EXPIRES_IN = 3600
const calcExpiry = () =>
  moment()
    .add(EXPIRES_IN, 'seconds')
    .format()

const decryptCredentias = (encryptedCredentials, password) => {
  const decreptedStr = decrypt(encryptedCredentials, password)
  const decreptedObj = JSON.parse(decreptedStr)
  return decreptedObj
}

const writeConfig = ({ domain, token, credentials, password }) => {
  // TODO: パスワードが指定されていなければ標準入力させる
  // TODO: tokenとcredentialsで関数を分けてそれぞれexports

  // トークンを保存
  token.expiry = calcExpiry()
  // TODO: ここはtokenだけ取得できればよいのでpassword不要
  const { token: oldToken } = readConfig({ domain, password })
  const domainDir = path.join(CONFIG_DIR, domain)
  fs.mkdirSync(domainDir, { recursive: true })
  console.log(path.join(domainDir, CREDENTIALS_FILE))
  fs.writeFileSync(path.join(domainDir, TOKEN_FILE), JSON.stringify({ ...oldToken, ...token }))

  // クレデンシャルを暗号化して保存
  const encryptedCredentials = encrypt(JSON.stringify(credentials), password)
  fs.writeFileSync(path.join(domainDir, CREDENTIALS_FILE), encryptedCredentials, 'utf-8')
}

const readConfig = ({ domain, password }) => {
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

module.exports = { writeConfig, readConfig }
