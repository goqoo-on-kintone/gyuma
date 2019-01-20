'use strict'

const express = require('express')
const fetch = require('node-fetch')
const qs = require('querystring')
const fs = require('fs')
const https = require('https')
const opener = require('opener')
require('dotenv').config()

const app = express()

// HTTPSサーバー起動
const options = {
  key: fs.readFileSync(process.env.HTTPS_KEY),
  cert: fs.readFileSync(process.env.HTTPS_CERT),
  passphrase: process.env.HTTPS_PASSPHRASE,
}
const httpsServer = https.createServer(options, app)

module.exports = params =>
  new Promise((resolve, reject) => {
    /* eslint-disable camelcase */
    const { clientId: client_id, clientSecret: client_secret, scope, domain, port = 3000 } = params
    const localhost = `https://localhost:${port}`
    const redirect_uri = `${localhost}/oauth2callback`
    const response_type = 'code'
    /* eslint-enable camelcase */
    const state = Math.random()
      .toString(36)
      .slice(-8)

    const authUri = `https://${domain}/oauth2/authorization`
    const tokenUri = `https://${domain}/oauth2/token`

    // 認可要求
    app.get('/', (req, res) => {
      const params = qs.stringify({
        client_id,
        redirect_uri,
        state,
        response_type,
        scope,
      })
      res.redirect(302, `${authUri}?${params}`)
    })

    // アクセストークンの要求・取得
    app.get('/oauth2callback', async (req, res) => {
      try {
        res.header('Content-Type', 'application/json; charset=utf-8')
        const resp = await fetch(tokenUri, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: qs.stringify({
            client_id,
            client_secret,
            code: req.query.code,
            grant_type: 'authorization_code',
            redirect_uri,
          }),
        }).then(res => res.json())

        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' })
        res.write('Authentication succeeded 🎉' + '</br>')
        res.end('Close the browser and return to the Terminal.')
        server.close()

        await resolve(resp)
      } catch (err) {
        await reject(err)
      }
    })

    const server = httpsServer.listen(port, () => {
      console.log(`listening on ${localhost}`)
      opener(localhost)
    })
  })
