---
sidebar_position: 5
---

# From Source

Clone the repository:

```bash
git clone https://github.com/yungwood/ical-filter-proxy.git
cd ical-filter-proxy
```

Run the tests:

```bash
go test ./...
```

Build the binary:

```bash
go build .
```

Run the service:

```bash
./ical-filter-proxy -config config.yaml
```
