# nuts-monitor
Application for monitoring a Nuts network. Used for security and health monitoring.

## Building and running
### Production
To build for production:

```shell
npm install
npm run dist
go run .
```

This will serve the front end from the embedded filesystem.
### Development

During front-end development, you probably want to use the real filesystem and webpack in watch mode:

```shell
npm install
npm run watch
go run . live
```

Generate APIs with:

```shell
make apis
```

### Test

There's a small test suite that can be run with

```shell
make test
```

### Docker
```shell
docker run -p 1313:1313 nutsfoundation/nuts-monitor
```

## Configuration
The location of the config file can be set using e cmdline flag or environment variable:

```shell
./monitor --configfile=./server.config.yaml
NUTS_CONFIGFILE=./server.config.yaml ./monitor
```

When running in Docker without a config file mounted at `/app/server.config.yaml` it will use the default configuration or you can change the command parameters.

The `nutsnodeapikeyfile` config parameter should point to a PEM encoded private key file. The corresponding public key should be configured on the Nuts node in SSH authorized keys format.
`nutsnodeapiuser` Is required when using Nuts node API token security. It must match the user in the SSH authorized keys file.
`nutsnodeapiaudience` must match the config parameter set in the Nuts node.
Check https://nuts-node.readthedocs.io for Nuts node API security details.

## Health check

The monitor exposes a status and health check endpoints on `/status` and `/health`. The health endpoint returns a sprint actuator style body.

## Technology Stack

Frontend framework is vue.js

Icons are from https://heroicons.com

CSS framework is https://tailwindcss.com