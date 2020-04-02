# FIXME: pin version
FROM golang:alpine AS build-go
# Note: if necessary, add bzr, mercurial, ... below
RUN apk add --no-cache git
ADD . .
RUN go build -o app

# FIXME: pin version ?
FROM alpine
RUN set -eux; \
        apk add --no-cache \
            ca-certificates \
            su-exec
WORKDIR /work
COPY --from=build-go app /work/
ENTRYPOINT su-exec 1000:1000 ./app
