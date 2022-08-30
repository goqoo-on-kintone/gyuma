export type ProxyOption =
  | string
  | {
      protocol: string
      auth: string
      hostname: string
      port: number
    }
