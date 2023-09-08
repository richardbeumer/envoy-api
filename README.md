# API for exposing Envoy Production values.

API retrieves token from enlighten and queries the local envoy for production stats

## Run local

To run loccaly perform the following steps
1. Create Python virtual environment and install the requirements in it.
```
python -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt --no-cache-dir
```

2. Set variables:

```
export ENLIGHTEN_USERNAME=<enlighten-username>
export ENLIGHTEN_PASSWORD=<enlighten-password>
export ENVOY_SERIAL=<envoy-serial>
export ENVOY_SITE=<envoy-siteId>
export ENVOY_HOST=<envoy-host-or-ip>
```

3. Run the code:
```
cd src
uvicorn api:app --reload
```

## DevContainer

You can also run from a [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

1. Start the devContainer. 
```
vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=<url-of-this-forked-repo>
```

2. Set variables:

```
export ENLIGHTEN_USERNAME=<enlighten-username>
export ENLIGHTEN_PASSWORD=<enlighten-password>
export ENVOY_SERIAL=<envoy-serial>
export ENVOY_SITE=<envoy-siteId>
export ENVOY_HOST=<envoy-host-or-ip>
```

3. Run the code from a terminal in the container.
```
cd src
uvicorn api:app --reload --host: 0.0.0.0
```