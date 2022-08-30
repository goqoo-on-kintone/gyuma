export type ServerParams = {
  domain: string
  scope: string
  password: string
  client_id: string
  client_secret: string
  port?: number
}

export type Client = Partial<{
  domain: string
  client_id: string
  client_secret: string
  state: string
}>

export type Query = Required<Client> & {
  scope: string
}

export type ProxyOption =
  | string
  | {
      protocol: string
      auth: string
      hostname: string
      port: number
    }
