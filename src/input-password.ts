import inquirer from 'inquirer'

const name = 'value' as const

type Question = Parameters<typeof inquirer.prompt>[number]
const singlePrompt = (question: Question) => inquirer.prompt<{ [name]: string }>([question]).then((_) => _[name])

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
    validate: async (firstInput) => {
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

export const inputPassword = async ({ exists }: { exists: boolean }) => {
  return exists ? inputExistingPassword() : inputNewPassword()
}

export const inputClientId = async () =>
  singlePrompt({
    message: 'Enter a Client ID',
    name,
  })

export const inputClientSecret = async () =>
  singlePrompt({
    type: 'password',
    message: 'Enter a Client Secret',
    name,
  })
