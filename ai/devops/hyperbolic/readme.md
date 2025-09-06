# Hyperbolic.ai

- https://hyperbolic.ai/
- https://app.hyperbolic.ai/
- https://docs.hyperbolic.xyz/

A great place to rent H100 VMs


## Setup - from local

### Hyperbolic one-time setup

> [!NOTE]
> Their cli or api seems broken, getting nothing but 404s when running their example commands
> You can skip this step for now, setup an instance in the web UI

https://docs.hyperbolic.xyz/docs/hyperbolic-cli

```sh
# install the cli
brew install HyperbolicLabs/hyperbolic/hyperbolic

hyperbolic auth login
```

### Starting an instance

Their cli or api seems broken, getting nothing but 404s when running their example commands

You can setup an instance in the UI though

Setup `.ssh/config`, adjust IP and Port

```
Host hyperbolic
  HostName 147.185.41.173
  User user
  IdentityFile ~/.ssh/id_ed25519
  Port 20013
```

## Setup - from vm

> [!WARNING]
> All subsequent commands are run from your Hyperbolic VM

`ssh hyperbolic`


### Setting up atmunge


```sh
git clone https://github.com/blebbit/atmunge
```

