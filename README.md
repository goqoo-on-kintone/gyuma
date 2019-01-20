# Gyuma OAuth
Gyuma is an OAuth client for kintone of Node.js.

## Install

### Local (for API)

```
$ npm install --save gyuma
 or
$ yarn add gyuma
```

### Global (for CLI)

```
$ npm install -g gyuma
 or
$ yarn global add gyuma
```


## API

```js
const gyuma = require('gyuma')
const request = require('request-promise')

;(async () => {
  const domain = 'example.cybozu.com'
  const params = {
    domain,
    client_id: process.env.OAUTH2_CLIENT_ID,
    client_secret: process.env.OAUTH2_CLIENT_SECRET,
    scope: 'k:app_settings:read k:app_settings:write',
    // port: 3000,
  }
  const { access_token } = await gyuma(params)
  const options = {
    url: `https://${domain}/k/v1/app/form/fields.json?app=256`,
    headers: {
      Authorization: `Bearer ${access_token}`,
    },
    json: true,
  }
  console.log(await request(options))
})()
```

## CLI

```
$ gyuma \
-d example.cybozu.com \
-i xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
-s xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
-S "k:app_settings:read k:app_settings:write"

$ curl --insecure https://localhost:3000/
```

### options
```
  -h, --help                          Output usage information
  -v, --version                       Output version information
  -d, --domain=<DOMAIN>               kintone domain name
  -i, --client_id=<CLIENT_ID>         kintone OAuth2 Client ID
  -s, --client_secret=<CLIENT_SECRET> kintone OAuth2 Client Secret
  -S, --scope=<SCOPE>                 kintone OAuth2 Scope
  -p, --port=<PORT>                   Web Server port number - defaults to 3000
```
