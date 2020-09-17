# ----------------------------------------------------------------------
FROM golang:1.15 AS build-srv

RUN mkdir -p /build/src
COPY src /build/src
WORKDIR /build/src
# Need to disable CGO because running the go app will fail due to
# different C stdlibs in the different docker image stages
# -- and we don't even need CGO...
#
# See https://forums.docker.com/t/standard-init-linux-go-175-exec-user-process-caused-no-such-file/20025/12
RUN CGO_ENABLED=0 go build -o /simple-spa-server


# ----------------------------------------------------------------------
FROM scratch

VOLUME [ "/data", "/data/conf", "/data/docroot" ]
COPY --from=build-srv /simple-spa-server /

CMD ["/simple-spa-server"]
