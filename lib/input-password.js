const inquirer = require('inquirer')

module.exports = async () =>
  inquirer
    .prompt([
      {
        type: 'password',
        message: 'Enter a password',
        name: 'val',
      },
    ])
    .then(_ => _.val)
