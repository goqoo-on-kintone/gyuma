'use strict'

const express = require('express')
const fetch = require('node-fetch')
const qs = require('querystring')
const fs = require('fs')
const https = require('https')
const opener = require('opener')
const path = require('path')
const del = require('del')
const moment = require('moment')
require('dotenv').config()
const createCertificate = require('./createCertificate')

const EXPIRES_IN = 3600

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
      // console.log('SSL Certificate is more than 30 days old. Removing.')

      del.sync([certPath], { force: true })

      certExists = false
    }
  }

  if (!certExists) {
    // console.log('Generating SSL Certificate')

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

const calcExpiry = () =>
  moment()
    .add(EXPIRES_IN, 'seconds')
    .format()

module.exports = (params, type) =>
  new Promise((resolve, reject) => {
    const tokens = {}
    const clients = {}
    let client = {}

    const { port = 3000 } = params
    const localhost = `https://localhost:${port}`
    const redirect_uri = `${localhost}/oauth2callback` // eslint-disable-line camelcase

    // 認可要求
    app.get('/oauth2', (req, res) => {
      const { domain, client_id, client_secret, scope } = req.query // eslint-disable-line camelcase
      const state = Math.random().toString(36)
      client = {
        domain,
        client_id,
        client_secret,
        state,
      }
      clients[domain] = client

      const params = qs.stringify({
        client_id,
        redirect_uri,
        state,
        scope,
        response_type: 'code',
      })
      const authUri = `https://${domain}/oauth2/authorization`
      res.redirect(302, `${authUri}?${params}`)
    })

    // アクセストークンの要求・保存
    app.get('/oauth2callback', async (req, res) => {
      try {
        res.header('Content-Type', 'text/html; charset=utf-8')

        // stateでCSRF攻撃を検証
        if (client.state !== req.query.state) {
          const code = 400
          const errorMessage = `${code} Invalid state`
          res.write(`<h1>${errorMessage} 💀</h1>`)
          res.status(code).end('<p>There is a suspicion of a CSRF attack!</p>')

          throw new Error(errorMessage)
        }

        const tokenUri = `https://${client.domain}/oauth2/token`
        const oauth2Res = await fetch(tokenUri, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: qs.stringify({
            client_id: client.client_id,
            client_secret: client.client_secret,
            redirect_uri,
            code: req.query.code,
            grant_type: 'authorization_code',
          }),
        })

        const resultMessage = '<p>Close the browser and return to the Terminal.</p>'

        if (!oauth2Res.ok) {
          const errorMessage = `${oauth2Res.status} ${oauth2Res.statusText}`
          res.write(`<h1>${errorMessage} ❌</h1>`)
          res.status(oauth2Res.status).end(resultMessage)

          throw new Error(errorMessage)
        }

        const token = await oauth2Res.json()
        token.expiry = calcExpiry()
        tokens[client.domain] = token

        res.write('<h1>Authentication succeeded 🎉</h1>')
        res.status(200).end(resultMessage)

        client = {}
        resolve(token)
      } catch (err) {
        reject(err)
      } finally {
        server.close()
      }
    })

    const server = httpsServer.listen(port, () => {
      // TODO: ブラウザ自動起動をしないマニュアルモードも作る
      // console.log(`listening on ${localhost}`)

      // ブラウザ自動起動 OAuthページを開く
      /* eslint-disable camelcase */
      const { domain, client_id, client_secret, scope } = params
      if (domain && client_id && client_secret && scope) {
        const uri = `${localhost}/oauth2?domain=${domain}&client_id=${client_id}&client_secret=${client_secret}&scope=${scope}`
        opener(encodeURI(uri))
      }
      /* eslint-enable camelcase */
    })
  })
