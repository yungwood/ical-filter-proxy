<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/yungwood/ical-filter-proxy">
    <img src="logo.png" alt="Logo" width="120" height="120">
  </a>

  <h3 align="center">iCal Filter Proxy</h3>

  <p align="center">
    iCal proxy with support for user-defined filtering rules
  </p>
</div>


## What's this thing?

Do you have iCal feeds with a bunch of stuff you *don't* need? Do you want to modify events generated by your rostering system?

iCal Filter Proxy is a simple service for proxying multiple iCal feeds while applying a list of filters to remove or modify events to suit your use case.

### Features

* Proxy multiple calendars
* Define a list of filters per calendar
* Match events using basic text and regex conditions
* Remove or modify events as they are proxied


### Built With

* Go
* [golang-ical](https://github.com/arran4/golang-ical)
* [yaml.v3](https://github.com/go-yaml/yaml/tree/v3.0.1)
* [DALL-E 2](https://openai.com/index/dall-e-2/) (app icon)


## Setup

### Docker

Docker images are published to [Docker Hub](https://hub.docker.com/r/yungwood/ical-filter-proxy). You'll need a config file (see below) mounted into the container at `/app/config.yaml`.

For example:

```bash
docker run -d \
  --name=ical-filter-proxy \
  -v config.yaml:/app/config.yaml \
  -p 8080:8080/tcp \
  --restart unless-stopped \
  yungwood/ical-filter-proxy:latest
```

You can also adapt the included [`docker-compose.yaml`](./docker-compose.yaml) example.

### Kubernetes

You can deploy iCal Filter Proxy using the included helm chart from [`charts/ical-filter-proxy`](charts/ical-filter-proxy).

### Build from source

You can also build the app and container from source.

```bash
# clone this repo
git clone git@github.com:yungwood/ical-filter-proxy.git
cd ical-filter-proxy

# build the ical-filter-proxy binary
go build .

# build container image
docker build \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg REVISION=$(git rev-parse HEAD) \
  --build-arg VERSION=$(git rev-parse --short HEAD) \
  -t ical-filter-proxy:latest .
```


## Configuration

Calendars and filters are defined in a yaml config file. By default this is `config.yaml` (use the `-config` switch to change this). The configuration must define at least one calendar for ical-filter-proxy to start.

Example configuration (with comments):

```yaml
calendars:

  # basic example
  - name: example # used as slug in URL - e.g. ical-filter-proxy:8080/calendars/example/feed?token=changeme
    token: "changeme" # token used to pull iCal feed - authentication is disabled when blank
    feed_url: "https://my-upstream-calendar.url/feed.ics" # URL for the upstream iCal feed
    filters: # optional - if no filters defined the upstream calendar is proxied as parsed
      - description: "Remove an event based on a regex"
        remove: true # events matching this filter will be removed
        match: # optional - all events will match if no rules defined
          summary: # match on event summary (title)
            contains: "deleteme" # must contain 'deleteme'
      - description: "Remove descriptions from all events"
        transform: # optional
          description: # modify event description
            remove: true # replace with a blank string
        
  # example: removing noise from an Office 365 calendar
  - name: outlook
    token: "changeme"
    feed_url: "https://outlook.office365.com/owa/calendar/.../reachcalendar.ics"
    filters:
      - description: "Remove canceled events" # canceled events remain with a 'Canceled:' prefix until removed
        remove: true
        match:
          summary:
            prefix: "Canceled: "
      - description: "Remove optional events"
        remove: true
        match:
          summary:
            prefix: "[Optional]"
      - description: "Remove public holidays"
        remove: true
        match:
          summary:
            regex_match: ".*[Pp]ublic [Hh]oliday.*"

  # example: cleaning up an OpsGenie feed
  - name: opsgenie
    token: "changeme"
    feed_url: "https://company.app.opsgenie.com/webapi/webcal/getRecentSchedule?webcalToken=..."
    filters:
      - description: "Keep oncall schedule events and fix names"
        match:
          summary:
            contains: "schedule: oncall"
        stop: true # stops processing any more filters
        transform:
          summary:
            replace: "On-Call" # replace the event summary (title)
      - description: "Remove all other events"
        remove: true

unsafe: false # optional - must be enabled if any calendars do not have a token
```


### Filters

Calendar events are filtered using a similar concept to email filtering. A list of filters is defined for each calendar in the config.

Each event parsed from `feed_url` is evaluated against the filters in sequence.

* All `match` rules for a filter must be true to match an event
* A filter with no `match` rules will *always* match
* When a match is found:
  * if `remove` is `true` the event is discarded
  * `transform` rules are applied to the event
  * if `stop` is `true` no more filters are processed
* If no match is found the event is retained by default

#### Match conditions

Each filter can spcify match conditions against the following event properties:

* `summary` (string value)
* `location` (string value)
* `description` (string value)

These match conditions are available for a string value:

* `contains` - property must contain this value
* `prefix` - property must start with this value
* `suffix` - property must end with this value
* `regex_match` - property must match the given regular expression (an invalid regex will result in no matches)

#### Transformations

Transformations can be applied to the following event properties:

* `summary` - string value
* `location` - string value
* `description` - string value

The following transformations are available for strings:

* `replace` - the property is replace with this value
* `remove` - if `true` the property is set to a blank string


## Roadmap to 1.0

There are a few more features I would like to add before I call the project "stable" and release version 1.0.

- [ ] Time based event conditions
- [ ] Caching
- [ ] Support for `ical_url_file` and `token_file` in config (vault secrets)
- [ ] Prometheus metrics
- [ ] Testing


## Contributing

If you have a suggestion that would make this better, please feel free to open an issue or send a pull request.


## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.


## Acknowledgments

This project was inspired by [darkphnx/ical-filter-proxy](https://github.com/darkphnx/ical-filter-proxy). I needed more flexibility with filtering rules and the ability to modify event descriptions... plus I wanted an excuse to finally write something in Go.
