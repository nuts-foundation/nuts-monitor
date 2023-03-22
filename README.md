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

### Test

There's a small test suite that can be run with

```shell
make test
```

### Docker
```shell
docker run -p 1323:1323 nutsfoundation/nuts-monitor
```

## Technology Stack

Frontend framework is vue.js

Icons are from https://heroicons.com

CSS framework is https://tailwindcss.com