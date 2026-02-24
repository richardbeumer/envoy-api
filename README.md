# API for exposing Envoy Production values.

API retrieves token from enlighten and queries the local envoy for production stats

## Run local

To run loccaly perform the following steps

1. Set variables:
```bash
export ENLIGHTEN_USERNAME=<enlighten-username>
export ENLIGHTEN_PASSWORD=<enlighten-password>
export ENVOY_SERIAL=<envoy-serial>
export ENVOY_SITE=<envoy-siteId>
export ENVOY_HOST=<envoy-host-or-ip>
```

2. Run the code:
```bash
cd app
go run main.go
```

## DevContainer

You can also run from a [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)


1. Start the devContainer. 
```bash
vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=<url-of-this-forked-repo>
```

2. Set variables:

```bash
export ENLIGHTEN_USERNAME=<enlighten-username>
export ENLIGHTEN_PASSWORD=<enlighten-password>
export ENVOY_SERIAL=<envoy-serial>
export ENVOY_SITE=<envoy-siteId>
export ENVOY_HOST=<envoy-host-or-ip>
```

3. Run the code from a terminal in the container.
```bash
cd app
go run main.go
```
