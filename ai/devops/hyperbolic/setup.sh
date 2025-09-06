#!/bin/bash
set -euo pipefail

sudo apt-get install -y \
  tree

# oh-my-zsh
# sudo apt-get install -y zsh
# sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
# omz theme set dst

# Install Go
# mkdir -p /tmp
# pushd /tmp
# wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
# sudo rm -rf /usr/local/go
# sudo tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz

echo 'PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.zshrc


# Install Docker
# ## Add Docker's official GPG key:
# sudo install -m 0755 -d /etc/apt/keyrings
# sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
# sudo chmod a+r /etc/apt/keyrings/docker.asc

# ## Add the repository to Apt sources:
# echo \
#   "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
#   $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | \
#   sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
# sudo apt-get update
# sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
# sudo usermod -aG docker $USER


# git config --global credential.helper store

# curl -fsSL https://ollama.com/install.sh | sh

# /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# echo >> /home/user/.zshrc
# echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/user/.zshrc
# eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

# brew install \
#   huggingface-cli \
#   uv

# huggingface-cli login

# echo "REBOOTING in 10s"
# sleep 10
# sudo reboot now
