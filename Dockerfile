FROM golang:1.13-alpine AS build-go
# Note: if necessary, add bzr, mercurial, ... below
RUN apk add --no-cache git
ADD . /src
RUN cd /src && go build -o app

# FIXME: pin version ?
# TODO: timezones stuff
FROM alpine
RUN set -eux; \
        apk add --no-cache \
            ca-certificates \
            su-exec
WORKDIR /work
COPY --from=build-go /src/app /work/
ENTRYPOINT mkdir -p /log && chown 1000:1000 /log && su-exec 1000:1000 ./app -rqlog /log/requests.log
