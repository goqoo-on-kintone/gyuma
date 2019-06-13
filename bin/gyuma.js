#!/usr/bin/env node

'use strict'

const minimist = require('minimist')
const main = require('../lib/main')

const trim = str => str.replace(/^\n|\n$/g, '')

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
  -p, --port=<PORT>                   Web Server port number - defaults to 3000
`)
  console.error(message)
  process.exit(returnCode)
}
const parseArgumentOptions = () => {
  const argv = minimist(process.argv.slice(2), {
    boolean: ['version', 'help', 'onetime'],
    string: ['domain', 'client_id', 'client_secret', 'scope', 'port'],
    alias: {
      v: 'version',
      h: 'help',
      d: 'domain',
      i: 'client_id',
      s: 'client_secret',
      S: 'scope',
      p: 'port',
    },
  })

  // TODO: 既存を読み取る場合は、client_id, client_secretは必須でなくす
  if (!argv.domain || !argv.client_id || !argv.client_secret || !argv.scope) {
    usageExit()
  }

  argv.scope = argv.scope.replace(/,/g, ' ')

  return argv
}

;(async () => {
  const argv = parseArgumentOptions()

  if (argv.version) {
    showVersion()
  } else if (argv.help) {
    usageExit(0)
  }

  try {
    await main(argv)
  } catch (err) {
    console.error(err)
  }
})()
