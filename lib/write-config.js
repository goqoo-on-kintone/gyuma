const fs = require('fs')
const path = require('path')
const { CONFIG_DIR, TOKEN_FILE } = require('./constants')
const readConfig = require('./read-config')

module.exports = ({ domain, token }) => {
  const { token: oldToken } = readConfig(domain)
  const domainDir = path.join(CONFIG_DIR, domain)
  fs.mkdirSync(domainDir, { recursive: true })
  fs.writeFileSync(path.join(domainDir, TOKEN_FILE), JSON.stringify({ ...oldToken, ...token }))
}
