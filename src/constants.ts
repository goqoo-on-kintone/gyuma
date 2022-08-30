import path from 'path'
import os from 'os'

// export const EXPIRES_IN = 3600
export const CONFIG_DIR = path.join(os.homedir(), '.config/gyuma')
export const TOKEN_FILE = 'token.json'
export const CREDENTIALS_FILE = 'credentials'
