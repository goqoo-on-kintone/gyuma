import express from 'express'
import fetch from 'node-fetch'
import qs from 'qs'
import fs from 'fs'
import https from 'https'
import opener from 'opener'
import path from 'path'
import del from 'del'
import dotenv from 'dotenv'
import { createCertificate } from './createCertificate'
import { createAgent } from './agent'
import { AgentOptions, Argv, Client, Query, Token } from './types'

dotenv.config()

// サーバー証明書の設定 存在しなければ自動生成
const options: https.ServerOptions = {}
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
    // @ts-expect-error
    if ((now - certStat.ctime) / certTtl > 30) {
      // console.log('SSL Certificate is more than 30 days old. Removing.')

      del.sync([certPath], { force: true })

      certExists = false
    }
  }

  if (!certExists) {
    // console.log('Generating SSL Certificate')

    const pems = createCertificate([{ name: 'commonName', value: 'localhost' }])

    fs.writeFileSync(certPath, pems.private + pems.cert, { encoding: 'utf-8' })
  }

  const fakeCert = fs.readFileSync(certPath)
  options.key = fakeCert
  options.cert = fakeCert
}

// HTTPSサーバー起動
const app = express()
const httpsServer = https.createServer(options, app)

export const server = (params: Argv): Promise<Token> =>
  new Promise((resolve, reject) => {
    const tokens: Record<string, Token> = {}
    const clients: Record<string, Client> = {}
    let client: Client = {}

    const { port = 3000, proxy, pfx } = params
    const localhost = `https://localhost:${port}`
    const redirect_uri = `${localhost}/oauth2callback`

    // 認可要求
    app.get('/oauth2', (req, res) => {
      const { domain, client_id, client_secret, scope } = req.query as Query
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
          agent: createAgent({ proxy, pfx }),
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

        const token = (await oauth2Res.json()) as Token
        tokens[client.domain!] = token

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
