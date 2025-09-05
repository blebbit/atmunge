#!/bin/bash
set -euo pipefail

sudo apt-get install -y \
  build-essential \
  htop \
  git


# zsh

git config --global credential.helper store

curl -fsSL https://ollama.com/install.sh | sh


/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

echo >> /home/user/.bashrc
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/user/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

brew install \
  huggingface-cli \
  uv

huggingface-cli login


pushd bento
uv sync