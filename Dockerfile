#
# Build frontend
#
FROM node:lts-alpine as frontend-builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY ./web/src ./web/src
COPY ./*.config.js .
RUN npm run dist

#
# Build backend
#
FROM golang:1.20.2-alpine as backend-builder

ARG TARGETARCH
ARG TARGETOS

LABEL maintainer="wout.slakhorst@nuts.nl"

RUN apk update \
 && update-ca-certificates

ENV GO111MODULE on
ENV GOPATH /

RUN mkdir /app && cd /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
COPY --from=frontend-builder /app/web/dist /app/web/dist
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /app/nuts-monitor

#
# Runtime
#
FROM alpine:3.17
RUN mkdir /app && cd /app
WORKDIR /app
COPY --from=backend-builder /app/nuts-monitor .
HEALTHCHECK --start-period=5s --timeout=5s --interval=5s \
    CMD wget --no-verbose --tries=1 --spider http://localhost:1323/status || exit 1
EXPOSE 1323
ENTRYPOINT ["/app/nuts-monitor"]
CMD ["--configfile","/app/server.config.yaml"]
