const server = require('./server')

module.exports = async params => {
  await server(params)
}
