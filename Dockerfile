FROM alpine:3.20.2

# build args and defaults
ARG BUILD_DATE="not-set"
ARG REVISION="unknown"
ARG VERSION="dev-build"

# set some labels
LABEL org.opencontainers.image.created="$BUILD_DATE" \
      org.opencontainers.image.base.name="alpine:3.20.2" \
      org.opencontainers.image.documentation="https://github.com/yungwood/ical-filter-proxy/tree/main/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/yungwood/ical-filter-proxy" \
      org.opencontainers.image.revision="$REVISION" \
      org.opencontainers.image.title="iCal Filter Proxy" \
      org.opencontainers.image.description="iCal proxy with support for user-defined filtering rules" \
      org.opencontainers.image.version="$VERSION"

# install dependencies
RUN apk --no-cache add gcompat=1.1.0-r4

# create a group and user
RUN addgroup -S icalfilterproxy && adduser -S -G icalfilterproxy icalfilterproxy

# switch to app user
USER icalfilterproxy

# set working dir
WORKDIR /app

# copy binary
COPY ical-filter-proxy /usr/bin/ical-filter-proxy

# expose port, define entrypoint
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/bin/ical-filter-proxy"]
CMD ["-config", "/app/config.yaml"]