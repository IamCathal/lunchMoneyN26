# lunchMoneyN26
Import and filter transaction data directly from [N26](https://n26.com) account into [lunch money](https://lunchmoney.app/). (100% experimental at the moment but security is my number one priority)

## Usage
```
  -d int
        Search for the last n days for any transactions (if not in webserver mode) (default 1)
  -webserver
        Run the application as a webserver
```
The application can either be run locally to manually execute the flow one request at a time or as a webserver that can be called from browser extension in the future to import transactions easily.

When running the application 'locally' both flags are required to be set. When running the application as a web server only the webserver flag needs to be set.

* To query the last `n` days for transactions not in webserver mode: `./lunchMoneyN26 -d n`

When running the application in webserver mode requests must be in the following form:
`POST http://[server's ip]:2944/transcations?days=n` where you want to search for transactions within the last `n` days. Starting the application in webserver mode only requires the `webserver` flag to be set.

* To start the application in web server mode: `./lunchMoneyN26 -webserver`

## Configuration
All of the following variables must be set in a `.env` file placed at the root of the project

| Variable     | Description |
| ----------- | ----------- |
| `N26_USERNAME`      | The email associated with your N26 account      |
| `N26_PASSWORD`   | Your password associated with your N26 account        |
| `N26_DEVICE_TOKEN`   | A random UUIDV4. Can easily be generated [here](https://www.uuidgenerator.net/)      |
| `LUNCHMONEY_TOKEN`   | API token to access your lunch money account which can be retrieved [here](https://my.lunchmoney.app/developers)      |
| `API_PORT`   | API port to be used when in webserver mode     |