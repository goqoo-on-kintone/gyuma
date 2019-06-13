const { isPast } = require('date-fns')
const readConfig = require('./read-config')
const server = require('./server')
const refresh = require('./refresh')
const writeConfig = require('./write-config')

module.exports = async argv => {
  const { domain } = argv
  const { token: oldToken, credentials: oldCredentials } = readConfig(domain)
  const params = { ...oldCredentials, ...argv }
  const credentials = { client_id: params.client_id, client_secret: params.client_secret }

  // トークン新規発行
  if (
    // 一度もトークン発行していない場合
    !oldToken.access_token ||
    // 発行済みトークンとscopeが違う場合
    oldToken.scope !== argv.scope
  ) {
    const token = await server(params)
    writeConfig({ domain, token, credentials })
    return token.access_token
  }

  // トークンの有効期限
  const expiry = new Date(oldToken.expiry)

  if (isPast(expiry)) {
    console.log('refresh')
    try {
      // トークン期限切れの場合はリフレッシュ
      const token = await refresh({ ...oldToken, ...params })
      writeConfig({ domain, token, credentials })
      return token.access_token
    } catch (e) {
      // リフレッシュも期限切れの場合は新規発行
      // TODO: ここでエラーが起きた時の対処を検討
      console.log('refresh failed.', e)
      const token = await server(params)
      writeConfig({ domain, token, credentials })
      return token.access_token
    }
  }

  // 期限内の場合はそのまま使う
  console.log('exists')
  return oldToken.access_token
}
