const fs = require('fs')
const path = require('path')
const moment = require('moment')
const { CONFIG_DIR, TOKEN_FILE, CREDENTIALS_FILE } = require('./constants')
const readConfig = require('./read-config')
const { encrypt } = require('./encrypt')

const EXPIRES_IN = 3600
const calcExpiry = () =>
  moment()
    .add(EXPIRES_IN, 'seconds')
    .format()

module.exports = ({ domain, token, credentials, password }) => {
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
