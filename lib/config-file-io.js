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

const writeToken = ({ domain, token }) => {
  token.expiry = calcExpiry()
  delete token.refresh_token

  const oldToken = readToken({ domain })
  const domainDir = path.join(CONFIG_DIR, domain)
  fs.mkdirSync(domainDir, { recursive: true })
  console.log(path.join(domainDir, CREDENTIALS_FILE))
  fs.writeFileSync(path.join(domainDir, TOKEN_FILE), JSON.stringify({ ...oldToken, ...token }))
}

const writeCredentials = ({ domain, token, credentials, password }) => {
  // TODO: パスワードが指定されていなければ標準入力させる
  const domainDir = path.join(CONFIG_DIR, domain)
  const encryptedCredentials = encrypt(JSON.stringify(credentials), password)
  fs.writeFileSync(path.join(domainDir, CREDENTIALS_FILE), encryptedCredentials, 'utf-8')
}

const writeConfig = ({ domain, token, credentials, password }) => {
  writeToken({ domain, token, credentials, password })
  writeCredentials({ domain, token, credentials, password })
}

const readToken = ({ domain }) => {
  const tokenFilePath = path.join(CONFIG_DIR, domain, TOKEN_FILE)
  const token = fs.existsSync(tokenFilePath) ? require(tokenFilePath) : {}
  return token
}

const readCredentials = ({ domain, password }) => {
  // TODO: パスワードが指定されていなければ標準入力させる
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  const encryptedCredentials = fs.existsSync(credentialsFilePath) && fs.readFileSync(credentialsFilePath, 'utf-8')
  const credentials = encryptedCredentials ? decryptCredentias(encryptedCredentials, password) : {}
  return credentials
}

const readConfig = ({ domain, password }) => ({
  token: readToken({ domain }),
  credentials: readCredentials({ domain, password }),
})

module.exports = {
  writeConfig,
  writeToken,
  writeCredentials,
  readConfig,
  readToken,
  readCredentials,
}
