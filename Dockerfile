#
# Build frontend
#
FROM node:lts-alpine as frontend-builder
WORKDIR /app
COPY package*.json ./
ENV NODE_ENV production
RUN npm install
COPY ./web ./web
COPY ./*.config.js .
RUN npm run dist

#
# Build backend
#
FROM golang:1.20.2-alpine as builder

ARG TARGETARCH
ARG TARGETOS

LABEL maintainer="wout.slakhorst@nuts.nl"

RUN apk update \
 && update-ca-certificates

ENV GO111MODULE on
ENV GOPATH /

RUN mkdir /opt/nuts && cd /opt/nuts
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /opt/nuts/monitor

#
# Runtime
#
FROM alpine:3.17.3
RUN apk update \
  && apk add --no-cache \
             tzdata \
             curl \
  && update-ca-certificates
COPY --from=builder /opt/nuts/monitor /usr/bin/monitor

HEALTHCHECK --start-period=30s --timeout=5s --interval=10s \
    CMD curl -f http://localhost:1323/status || exit 1

EXPOSE 1323
ENTRYPOINT ["/usr/bin/monitor"]
CMD ["--configfile","/app/server.config.yaml"]
