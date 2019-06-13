const fetch = require('node-fetch')
const qs = require('querystring')
const moment = require('moment')

const EXPIRES_IN = 3600

const calcExpiry = () =>
  moment()
    .add(EXPIRES_IN, 'seconds')
    .format()

// アクセストークンのリフレッシュ
/* eslint-disable camelcase */
module.exports = async ({ domain, client_id, client_secret, refresh_token }) => {
  const tokenUri = `https://${domain}/oauth2/token`
  const oauth2Res = await fetch(tokenUri, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: qs.stringify({
      client_id,
      client_secret,
      refresh_token,
      grant_type: 'refresh_token',
    }),
  })

  if (!oauth2Res.ok) {
    const errorMessage = `${oauth2Res.status} ${oauth2Res.statusText}`
    throw new Error(errorMessage)
  }
  const token = await oauth2Res.json()
  token.expiry = calcExpiry()
  return token
}
