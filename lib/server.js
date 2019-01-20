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

let token
module.exports = params =>
  new Promise((resolve, reject) => {
    /* eslint-disable camelcase */
    const { client_id, client_secret, scope, domain, port = 3000, onetime = true } = params
    const localhost = `https://localhost:${port}`
    const redirect_uri = `${localhost}/oauth2callback`
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
        scope,
        response_type: 'code',
      })
      res.redirect(302, `${authUri}?${params}`)
    })

    // アクセストークンの要求・保存
    app.get('/oauth2callback', async (req, res) => {
      try {
        res.header('Content-Type', 'application/json; charset=utf-8')
        const oauthRes = await fetch(tokenUri, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: qs.stringify({
            client_id,
            client_secret,
            redirect_uri,
            code: req.query.code,
            grant_type: 'authorization_code',
          }),
        })

        const resultHeader = { 'Content-Type': 'text/html; charset=utf-8' }
        const resultMessage = '<p>Close the browser and return to the Terminal.</p>'

        if (!oauthRes.ok) {
          const errorMessage = `${oauthRes.status} ${oauthRes.statusText}`
          res.writeHead(oauthRes.status, resultHeader)
          res.write(`<h1>${errorMessage} ❌</h1>`)
          res.end(resultMessage)

          throw new Error(errorMessage)
        }

        res.writeHead(200, resultHeader)
        res.write('<h1>Authentication succeeded 🎉</h1>')
        res.end(resultMessage)
        token = await oauthRes.json()
        const now = moment()
        token.createdAt = now
        token.updatedAt = now
        console.log('token createdAt: ' + now.format())

        resolve(token)
      } catch (err) {
        reject(err)
      } finally {
        if (onetime) {
          server.close()
        }
      }
    })

    // アクセストークンのリフレッシュ
    const refreshToken = async () => {
      try {
        const oauthRes = await fetch(tokenUri, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: qs.stringify({
            client_id,
            client_secret,
            refresh_token: token.refresh_token,
            grant_type: 'refresh_token',
          }),
        })

        if (!oauthRes.ok) {
          const errorMessage = `${oauthRes.status} ${oauthRes.statusText}`
          throw new Error(errorMessage)
        }

        const now = moment()
        token = { ...token, ...(await oauthRes.json()), updatedAt: now }
        console.log('token updatedAt: ' + now.format())
      } catch (err) {
        reject(err)
      }
    }

    // 保存済みトークンの取得
    app.get('/token', (req, res) => {
      console.log(token)
      res.header('Content-Type', 'application/json; charset=utf-8')
      res.send(token)
    })

    const server = httpsServer.listen(port, () => {
      console.log(`listening on ${localhost}`)
      opener(localhost)

      if (onetime) {
        return
      }

      // タイマー監視 アクセストークンの期限切れが近づいたらリフレッシュ
      const INTERVAL_MINUTES = 1
      const LIMIT_MINUTES = 55
      setInterval(() => {
        if (!token) {
          return
        }
        const { expires_in: expiresIn } = token
        const diff = moment().diff(token.updatedAt, 'seconds')
        console.log(diff)
        if (expiresIn - diff < LIMIT_MINUTES * 60) {
          refreshToken()
        }
      }, INTERVAL_MINUTES * 60 * 1000)
    })
  })
