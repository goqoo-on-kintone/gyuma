#!/usr/bin/env node

import minimist from 'minimist'
import { gyuma } from './main'
import { Argv } from './types'

const trim = (str: string) => str.replace(/^\n|\n$/g, '')

const showVersion = () => {
  const { version } = require('../package.json')
  console.error(`Gyuma OAuth ${version}`)
  process.exit(0)
}

const usageExit = (returnCode = 0) => {
  const message = trim(`
usage: gyuma [<options>]

  -h, --help                          Output usage information
  -v, --version                       Output version information
  -d, --domain=<DOMAIN>               kintone domain name
  -i, --client_id=<CLIENT_ID>         kintone OAuth2 Client ID
  -s, --client_secret=<CLIENT_SECRET> kintone OAuth2 Client Secret
  -S, --scope=<SCOPE>                 kintone OAuth2 Scope
  -p, --password=<PASSWORD>           Credentials Password
  -P, --port=<PORT>                   Web Server port number - defaults to 3000
      --proxy=<PROXY>                 Proxy Server
      --pfxFilepath=<PFX-FILEPATH>    Path to client certificate file
      --pfxPassword=<PFX-PASSWORD>    Password of client certificate
`)
  console.error(message)
  process.exit(returnCode)
}
const parseArgumentOptions = () => {
  const argv = minimist(process.argv.slice(2), {
    boolean: ['version', 'help', 'onetime', 'noprompt'],
    string: ['domain', 'client_id', 'client_secret', 'scope', 'port', 'proxy', 'pfxFilepath', 'pfxPassword'],
    alias: {
      v: 'version',
      h: 'help',
      d: 'domain',
      i: 'client_id',
      s: 'client_secret',
      S: 'scope',
      p: 'password',
      P: 'port',
    },
  })

  if (!argv.domain || !argv.scope) {
    usageExit()
  }

  argv.scope = argv.scope.replace(/,/g, ' ')

  const { pfxFilepath, pfxPassword, ...ret } = argv
  ret.pfx = {
    filepath: pfxFilepath,
    password: pfxPassword,
  }

  return ret
}

;(async () => {
  const argv = parseArgumentOptions()

  if (argv.version) {
    showVersion()
  } else if (argv.help) {
    usageExit(0)
  }

  try {
    const CLI = true
    const accessToken = await gyuma(argv as unknown as Argv, CLI)
    console.info(accessToken)
  } catch (err) {
    console.error(err)
  }
})()
