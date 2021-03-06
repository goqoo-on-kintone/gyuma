const inquirer = require('inquirer')

const name = 'value'
const singlePrompt = question => inquirer.prompt([question]).then(_ => _[name])

const inputExistingPassword = async () =>
  singlePrompt({
    type: 'password',
    message: 'Enter a password',
    name,
  })

const inputNewPassword = async () =>
  singlePrompt({
    type: 'password',
    message: 'Enter new Gyuma password',
    name,
    validate: async firstInput => {
      const secondInput = await confirmNewPassword()
      if (firstInput === secondInput) {
        return true
      }
      return 'Passwords do not match'
    },
  })

const confirmNewPassword = async () =>
  singlePrompt({
    type: 'password',
    message: 'Retype new Gyuma password',
    name,
  })

module.exports = async ({ exists }) => {
  return exists ? inputExistingPassword() : inputNewPassword()
}
