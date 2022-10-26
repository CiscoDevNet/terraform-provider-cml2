#!/usr/bin/env bash

# set -x

# check if we have everything...
if ! which ngrok gh curl tmux mkkey; then
    echo "required command missing!"
    exit
fi

REPO="ciscodevnet/terraform-provider-cml2"
CML="https://cml-controller.cml.lab:443"

# check if ngrok is running
if ! curl >/dev/null -sf localhost:4040/api; then
    echo "starting tmux and ngrok"
    if ! tmux has-session; then
        tmux new-session -d
    fi
    tmux new-window  -n "ngrok" ngrok start --none
    sleep 1
    if ! >/dev/null curl -sf localhost:4040/api; then
        echo "can't start ngrok, failing"
        exit 1
    else
        echo "tmux started, ngrok started"
    fi
fi

# get the tunnel from the agent and start it, if no tunnel
TUNNEL=$(curl -sf localhost:4040/api/tunnels | jq -r '.tunnels|map(select(.config.addr == "'$CML'"))[0]|.public_url')
if [ "$TUNNEL" = "null" ]; then
    DATA='{"proto": "http","addr": "'$CML'","name": "cml"}'
    TUNNEL=$(echo $DATA | curl -sf -XPOST -d@- -H "Content-Type: application/json" localhost:4040/api/tunnels | jq -r '.public_url')
fi

# read the public github key for our repo
read -d' ' KEY_ID KEY <<< "$(gh api /repos/$REPO/actions/secrets/public-key | jq -r '.|.key_id, .key')"

# {
#   "key_id": "012345678912345678",
#   "key": "2Sg8iYjAxxmI2LvUXpJjkYrMxURPc8r+dB7TJyvv1234"
# }

# create the encrypted secret from our tunnel endpoint URL
export GH_KEY="$KEY"
MSG=$(~/go/bin/mkkey $TUNNEL)

# create/update the secret on github (NGROK_URL is the secret name)
echo '{"encrypted_value":"'$MSG'","key_id":"'$KEY_ID'"}' | \
gh api -XPUT /repos/$REPO/actions/secrets/NGROK_URL --input -
