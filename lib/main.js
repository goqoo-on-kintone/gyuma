const { isPast } = require('date-fns')
const server = require('./server')
const { writeToken, writeCredentials, readToken, readCredentials, existsCredentialsFile } = require('./config-file-io')
const inputPassword = require('./input-password')

const generateNewAccessToken = async argv => {
  if (!argv.password && argv.noprompt) {
    throw new Error('password is not found!')
  }

  const { domain } = argv
  const password = argv.password || (await inputPassword({ exists: existsCredentialsFile({ domain }) }))
  const oldCredentials = readCredentials({ domain, password })
  const params = { ...oldCredentials, ...argv }
  const credentials = { client_id: params.client_id, client_secret: params.client_secret }

  const token = await server(params)
  writeToken({ domain, token })
  writeCredentials({ domain, credentials, password })

  return token.access_token
}

module.exports = async (argv, CLI = false) => {
  if (!CLI) {
    argv.noprompt = true
  }

  const oldToken = readToken({ domain: argv.domain })

  if (!oldToken.access_token) {
    // 一度もトークン発行していない場合
    return generateNewAccessToken(argv)
  }

  if (oldToken.scope !== argv.scope) {
    // 発行済みトークンとscopeが違う場合
    return generateNewAccessToken(argv)
  }

  if (isPast(new Date(oldToken.expiry))) {
    // トークンが期限切れの場合
    return generateNewAccessToken(argv)
  }

  // 同一scopeで期限内のトークンが残っていれば、そのまま使う
  return oldToken.access_token
}
