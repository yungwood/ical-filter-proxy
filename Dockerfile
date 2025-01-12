# Stage 1: go build
FROM golang:1.23-alpine as builder

# build args and defaults
ARG VERSION="unknown"

# set working dir
WORKDIR /src

# install build tools
RUN apk --no-cache add build-base

# download go module deps
COPY go.mod go.sum ./
RUN go version
RUN go mod download

# build binary
COPY . .
RUN go build -ldflags "-X 'main.version=${VERSION}'" -o /usr/bin/ical-filter-proxy .


# Stage 2: docker image
FROM alpine:3.20.2

# install dependencies
RUN apk --no-cache add gcompat=1.1.0-r4

# create a group and user
RUN addgroup -S icalfilterproxy && adduser -S -G icalfilterproxy icalfilterproxy

# set working dir
WORKDIR /app

# copy binary
COPY --from=builder /usr/bin/ical-filter-proxy /usr/bin/ical-filter-proxy

# switch to app user
USER icalfilterproxy

# expose port, define entrypoint
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/bin/ical-filter-proxy"]
CMD ["-config", "/app/config.yaml"]
