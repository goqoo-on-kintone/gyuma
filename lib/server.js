'use strict'

const express = require('express')
const fetch = require('node-fetch')
const qs = require('querystring')
const fs = require('fs')
const https = require('https')
const opener = require('opener')
const path = require('path')
const del = require('del')
require('dotenv').config()
const createCertificate = require('./createCertificate')

// サーバー証明書の設定 存在しなければ自動生成
const options = {}
if (process.env.HTTPS_KEY && process.env.HTTPS_CERT) {
  options.key = fs.readFileSync(process.env.HTTPS_KEY)
  options.cert = fs.readFileSync(process.env.HTTPS_CERT)
  if (process.env.HTTPS_PASSPHRASE) {
    options.passphrase = process.env.HTTPS_PASSPHRASE
  }
} else {
  // Use a self-signed certificate if no certificate was configured.
  // Cycle certs every 24 hours
  const certPath = path.join(__dirname, '../ssl/server.pem')

  let certExists = fs.existsSync(certPath)

  if (certExists) {
    const certTtl = 1000 * 60 * 60 * 24
    const certStat = fs.statSync(certPath)

    const now = new Date()

    // cert is more than 30 days old, kill it with fire
    if ((now - certStat.ctime) / certTtl > 30) {
      console.log('SSL Certificate is more than 30 days old. Removing.')

      del.sync([certPath], { force: true })

      certExists = false
    }
  }

  if (!certExists) {
    console.log('Generating SSL Certificate')

    const attrs = [{ name: 'commonName', value: 'localhost' }]

    const pems = createCertificate(attrs)

    fs.writeFileSync(certPath, pems.private + pems.cert, { encoding: 'utf-8' })
  }

  const fakeCert = fs.readFileSync(certPath)
  options.key = fakeCert
  options.cert = fakeCert
}

// HTTPSサーバー起動
const app = express()
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
