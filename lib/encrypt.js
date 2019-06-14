const crypto = require('crypto')

const ALGORITHM = 'aes-256-cbc'
const SALT = 'lKzR+i6IwG/DbuAY5thksw=='
const ENCODING = 'base64'

// 暗号化メソッド
const encrypt = (rawData, password) => {
  // 鍵、IV、暗合器を生成
  const key = crypto.scryptSync(password, SALT, 32)
  const iv = crypto.randomBytes(16)
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv)

  // 暗号化
  let encryptedData = cipher.update(Buffer.from(rawData))
  encryptedData = Buffer.concat([encryptedData, cipher.final()])

  const ciphertext = Buffer.concat([iv, encryptedData])
  const encodedCiphertext = ciphertext.toString(ENCODING)
  return encodedCiphertext
}

// 複合メソッド
const decrypt = (encodedCiphertext, password) => {
  const ciphertext = Buffer.from(encodedCiphertext, ENCODING)

  // iv is first 16 bytes of the ciphertext
  const iv = ciphertext.slice(0, 16)
  const encryptedData = ciphertext.slice(16)

  // 鍵、暗合器を生成
  const key = crypto.scryptSync(password, SALT, 32)
  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv)

  try {
    // 複合
    let decryptedData = decipher.update(encryptedData)
    decryptedData = Buffer.concat([decryptedData, decipher.final()])

    const rawData = decryptedData.toString()
    return rawData
  } catch (e) {
    if (e.message.includes('EVP_DecryptFinal_ex')) {
      console.error('Failed to decrypt credentials file!')
      process.exit(1)
    }
  }
}

module.exports = {
  encrypt,
  decrypt,
}
