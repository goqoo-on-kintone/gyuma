export type Argv = {
  domain: string
  scope: string
  password?: string
  client_id?: string
  client_secret?: string
  port?: number

  proxy?: ProxyOption
  pfx?: PfxOption

  noprompt?: boolean
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

export type Credentials = {
  client_id: string
  client_secret: string
}

export type Token = {
  expiry: string
  refresh_token?: string
  access_token: string
  scope: string
}

export type ProxyOption =
  | string
  | {
      protocol: string
      auth?: string
      hostname: string
      port: number
    }
export type PfxOption = { filepath: string; password: string }
export type AgentOptions = { proxy?: ProxyOption; pfx?: PfxOption }
