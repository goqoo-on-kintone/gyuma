const { isPast } = require('date-fns')
const readConfig = require('./read-config')
const server = require('./server')
const refresh = require('./refresh')
const writeConfig = require('./write-config')

module.exports = async argv => {
  const { domain } = argv
  const { token: oldToken, credentials } = readConfig(domain)
  const params = { ...credentials, ...argv }

  // 一度もトークン発行していない場合
  if (!oldToken.access_token) {
    const token = await server(params)
    writeConfig({ domain, token })
    return token.access_token
  }

  // トークンの有効期限
  const expiry = new Date(oldToken.expiry)

  // 期限切れの場合はリフレッシュ
  if (isPast(expiry)) {
    console.log('refresh')
    try {
      const token = await refresh({ ...oldToken, ...params })
      writeConfig({ domain, token })
      return token.access_token
    } catch (e) {
      console.log('refresh failed.', e)
      const token = await server(params)
      writeConfig({ domain, token })
      return token.access_token
    }
  }

  // 期限内の場合はそのまま使う
  console.log('exists')
  return oldToken.access_token
}
