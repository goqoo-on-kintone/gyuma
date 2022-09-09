import fs from 'fs'
import path from 'path'
import moment from 'moment'
import { CONFIG_DIR, TOKEN_FILE, CREDENTIALS_FILE } from './constants'
import { encrypt, decrypt } from './encrypt'
import { Credentials, Token } from './types'

const EXPIRES_IN = 3600
const calcExpiry = () => moment().add(EXPIRES_IN, 'seconds').format()

const decryptCredentias = (encryptedCredentials: string, password: string) => {
  const decreptedStr = decrypt(encryptedCredentials, password)
  const decreptedObj = JSON.parse(decreptedStr!)
  return decreptedObj
}

export const writeToken = ({ domain, token }: { domain: string; token: Token }) => {
  token.expiry = calcExpiry()
  delete token.refresh_token

  const oldToken = readToken({ domain })
  const domainDir = path.join(CONFIG_DIR, domain)
  fs.mkdirSync(domainDir, { recursive: true })
  console.info(path.join(domainDir, CREDENTIALS_FILE))
  fs.writeFileSync(path.join(domainDir, TOKEN_FILE), JSON.stringify({ ...oldToken, ...token }))
}

// TODO: credentialsは1ファイルのみに変更、ファイル内に複数ドメインの情報を保存
// 今のままではドメインごとにパスワードを別途作成する必要があって辛いので
export const writeCredentials = ({
  domain,
  credentials,
  password,
}: {
  domain: string
  credentials: Credentials
  password: string
}) => {
  const domainDir = path.join(CONFIG_DIR, domain)
  const encryptedCredentials = encrypt(JSON.stringify(credentials), password)
  fs.writeFileSync(path.join(domainDir, CREDENTIALS_FILE), encryptedCredentials, 'utf-8')
}

export const readToken = ({ domain }: { domain: string }): Token => {
  const tokenFilePath = path.join(CONFIG_DIR, domain, TOKEN_FILE)
  const token = fs.existsSync(tokenFilePath) ? require(tokenFilePath) : {}
  return token
}

// TODO: credentialsは1ファイルのみにした後は、該当ドメインの情報のみreturnする
export const readCredentials = ({ domain, password }: { domain: string; password: string }) => {
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  const encryptedCredentials = fs.existsSync(credentialsFilePath) && fs.readFileSync(credentialsFilePath, 'utf-8')
  const credentials = encryptedCredentials ? decryptCredentias(encryptedCredentials, password) : {}
  return credentials
}

export const existsCredentialsFile = ({ domain }: { domain: string }) => {
  const credentialsFilePath = path.join(CONFIG_DIR, domain, CREDENTIALS_FILE)
  return fs.existsSync(credentialsFilePath)
}
