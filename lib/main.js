const { isPast } = require('date-fns')
const server = require('./server')
const { writeConfig, readConfig } = require('./config-file-io')

const generateNewAccessToken = async ({ params, domain, credentials, password }) => {
  const token = await server(params)
  writeConfig({ domain, token, credentials, password })
  return token.access_token
}

module.exports = async argv => {
  const { domain, password } = argv
  const { token: oldToken, credentials: oldCredentials } = readConfig({ domain, password })
  const params = { ...oldCredentials, ...argv }
  const credentials = { client_id: params.client_id, client_secret: params.client_secret }

  if (!oldToken.access_token) {
    // 一度もトークン発行していない場合
    return generateNewAccessToken({ params, domain, credentials, password })
  }

  if (oldToken.scope !== argv.scope) {
    // 発行済みトークンとscopeが違う場合
    return generateNewAccessToken({ params, domain, credentials, password })
  }

  if (isPast(new Date(oldToken.expiry))) {
    // トークンが期限切れの場合
    return generateNewAccessToken({ params, domain, credentials, password })
  }

  // 同一scopeで期限内のトークンが残っていれば、そのまま使う
  return oldToken.access_token
}
