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

// TODO: credentialsは1ファイルのみに変更、ファイル内に複数ドメインの情報を保存
// 今のままではドメインごとにパスワードを別途作成する必要があって辛いので
const writeCredentials = ({ domain, credentials, password }) => {
  const domainDir = path.join(CONFIG_DIR, domain)
  const encryptedCredentials = encrypt(JSON.stringify(credentials), password)
  fs.writeFileSync(path.join(domainDir, CREDENTIALS_FILE), encryptedCredentials, 'utf-8')
}

const readToken = ({ domain }) => {
  const tokenFilePath = path.join(CONFIG_DIR, domain, TOKEN_FILE)
  const token = fs.existsSync(tokenFilePath) ? require(tokenFilePath) : {}
  return token
}

// TODO: credentialsは1ファイルのみにした後は、該当ドメインの情報のみreturnする
const readCredentials = ({ domain, password }) => {
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  const encryptedCredentials = fs.existsSync(credentialsFilePath) && fs.readFileSync(credentialsFilePath, 'utf-8')
  const credentials = encryptedCredentials ? decryptCredentias(encryptedCredentials, password) : {}
  return credentials
}

const existsCredentialsFile = ({ domain }) => {
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  return fs.existsSync(credentialsFilePath)
}

module.exports = {
  writeToken,
  writeCredentials,
  readToken,
  readCredentials,
  existsCredentialsFile,
}
