const { isPast } = require('date-fns')
const { server } = require('./server')
const { writeToken, writeCredentials, readToken, readCredentials, existsCredentialsFile } = require('./config-file-io')
const { inputPassword, inputClientId, inputClientSecret } = require('./input-password')

const generateNewAccessToken = async (argv) => {
  if (!argv.domain) {
    throw new Error('domain is not found!')
  }
  const { domain } = argv

  if (!argv.password) {
    if (argv.noprompt) throw new Error('password is not found!')
    else argv.password = await inputPassword({ exists: existsCredentialsFile({ domain }) })
  }
  const { password } = argv

  const oldCredentials = readCredentials({ domain, password })
  argv.client_id ||= oldCredentials.client_id
  argv.client_secret ||= oldCredentials.client_secret
  if (!argv.client_id) {
    if (argv.noprompt) throw new Error('client_id is not found!')
    else argv.client_id = await inputClientId()
  }
  if (!argv.client_secret) {
    if (argv.noprompt) throw new Error('client_secret is not found!')
    else argv.client_secret = await inputClientSecret()
  }
  const credentials = { client_id: argv.client_id, client_secret: argv.client_secret }

  const token = await server({ ...argv, ...credentials })
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
