const fs = require('fs')
const path = require('path')
const moment = require('moment')
const { CONFIG_DIR, TOKEN_FILE } = require('./constants')
const readConfig = require('./read-config')

const EXPIRES_IN = 3600
const calcExpiry = () =>
  moment()
    .add(EXPIRES_IN, 'seconds')
    .format()

// TODO: credentialsを保存
// TODO: credentialsを暗号化

module.exports = ({ domain, token }) => {
  token.expiry = calcExpiry()
  const { token: oldToken } = readConfig(domain)

  const domainDir = path.join(CONFIG_DIR, domain)
  fs.mkdirSync(domainDir, { recursive: true })
  fs.writeFileSync(path.join(domainDir, TOKEN_FILE), JSON.stringify({ ...oldToken, ...token }))
}
