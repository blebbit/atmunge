#!/bin/bash
set -euo pipefail

${GPU:-t4}
${GPU_COUNT:-4}

# create vm
gcloud compute instances create \
  atmunge-gpu \
  --zone=us-central1-f \
  --machine-type=n1-highmem-4 \
  --accelerator=type=${GPU},count=${GPU_COUNT} \
  --image-family=debian-12-bookworm-v20250812 \
  --image-project=debian-12 \
  --boot-disk-size=200GB

# wait for it to come up

# copy overlay
gcloud compute scp --recurse ./overlay/* atmunge-gpu:~/

# copy bento
gcloud compute scp --recurse ../bento/* atmunge-gpu:~/bento/

# run setup
gcloud compute ssh atmunge-gpu -- 'bash -s' < ./overlay/setup.sh

# run test
# gcloud compute ssh atmunge-gpu -- 'bash -s' < ./overlay/test-curl.sh
