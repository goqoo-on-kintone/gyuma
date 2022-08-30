export type ServerParams = {
  domain: string
  scope: string
  password: string
  client_id: string
  client_secret: string
  port?: number
}

export type Options = Partial<{
  key: Buffer
  cert: Buffer
  passphrase: string
}>

export type Query = Partial<{
  domain: any
  client_id: any
  client_secret: any
  state: any
  scope: any
}>

export type Client = Partial<{
  domain: any
  client_id: any
  client_secret: any
  state: any
}>

export type ProxyOption =
  | string
  | {
      protocol: string
      auth: string
      hostname: string
      port: number
    }
