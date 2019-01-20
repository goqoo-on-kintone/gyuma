#!/usr/bin/env node

'use strict'

const minimist = require('minimist')
const server = require('../lib/server')

const parseArgumentOptions = () => {
  const argv = minimist(process.argv.slice(2), {
    boolean: ['version', 'help', 'onetime'],
    string: ['client_id', 'client_secret', 'scope', 'domain', 'port'],
    alias: {
      v: 'version',
      h: 'help',
      i: 'client_id',
      s: 'client_secret',
      S: 'scope',
      d: 'domain',
      p: 'port',
    },
  })

  return argv
}

const main = async () => {
  const argv = parseArgumentOptions()
  try {
    await server(argv)
  } catch (err) {
    console.error(err)
  }
}

main()
