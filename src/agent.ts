import ProxyAgent from 'proxy-agent'
import https from 'https'
import fs from 'fs'
import type { PfxOption, ProxyOption } from './types'

// @ts-expect-error
const createProxyAgent = (proxy?: ProxyOption) => (proxy ? new ProxyAgent(proxy) : undefined)
const createPfxAgent = (pfx?: PfxOption) =>
  pfx && pfx.pfxFilepath && pfx.pfxPassword
    ? new https.Agent({
        pfx: fs.readFileSync(pfx.pfxFilepath),
        passphrase: pfx.pfxPassword,
      })
    : undefined

export const createAgent = ({ proxy, pfx }: { proxy?: ProxyOption; pfx?: PfxOption } = {}) => {
  const proxyAgent = createProxyAgent(proxy)
  const pfxAgent = createPfxAgent(pfx)

  if (proxyAgent && pfxAgent) {
    throw new Error('Proxy server and client certificate cannot be used at the same time.')
  }

  return proxyAgent ?? pfxAgent
}
