# Hyperbolic.ai

- https://hyperbolic.ai/
- https://app.hyperbolic.ai/
- https://docs.hyperbolic.xyz/

A great place to rent H100 VMs



## Setup - from your computer

You should already have a

- huggingface account, you will need a create token during setup
  - some models require you accept terms before being able to download too (i.e. all of the safety ones)
- hyperbolic account, because that's what we're doing after all

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
# get the repo
git clone https://github.com/blebbit/atmunge
cd atmunge

# more vm setup
./ai/devops/hyperbolic/setup.sh

# ...
# there will be several prompts, a reboot, and a relog
# ...

# build the command
go install ./cmd/atmunge

# test it works
atmunge firehose

# setup python
uv sync
```

## Running Models

We mainly use bentoml to do these things

There are single model servers and multi-model servers.
There are lots of interesting ways to compose them
and bentoml makes that super easy
while still giving you low-level access to the models
if you need it.

In any of the subdirs, run

```sh
cd ./ai/bento <model>
uv run bentoml serve
```

| Model | class | inputs | outputs | notes |
|:----|-|-|-|
| shieldgemma  | safety | text  | policy score | custom policy |
| shieldgemma2 | safety | image | scores | custom policy |
| llamaguard4  | safety | text & image | boolean | custom policy |
| safety-3x    | safety | all of the above | mixed | as above |

Models to be added

- https://huggingface.co/zentropi-ai/cope-a-9b
- gemma3-27b-it
- 

- https://github.com/huggingface/transformers/blob/main/docs/source/en/model_doc/gemma3.md
- https://github.com/huggingface/transformers/blob/main/docs/source/en/model_doc/shieldgemma2.md

> [!NOTE]
> note to self, write way more about the models, in a separate page(s)

### Testing

Then you can curl, body depending on the model

```sh
time curl -X 'POST' \
  'http://localhost:3000/check' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "..."
  }'
```

