#!/usr/bin/env bash

# set -x

# gh CLI from https://cli.github.com/
# ghsecret from https://github.com/rschmied/ghsecret
#
# change these two to local requirements:

REPO="ciscodevnet/terraform-provider-cml2"
CML="https://cml-controller.cml.lab:443"

# nothing to be changed further down

function help() {
  cmd=$(basename $0)
  cat <<EOT
$cmd usage:

$cmd start -- starts ngrok in tmux and provisions credentials to GH
$cmd stop -- stops tmux (and ngrok) and removes credentials from GH
$cmd force -- forcefully removes credentials from GH
$cmd status -- shows the status (also the default)
$cmd open -- opens the tmux session
$cmd -h | --help | help -- shows this help

Requirements:
- TF_VAR_username and TF_VAR_password environment variables with CML credentials
- TF_VAR_token as an alternative (and username / password empty)
- authorized gh tool (Github cli)
- curl, ghsecret, jq, ngrok and tmux in the path
- ngrok authtoken provided via ~/.ngrok2/ngrok.yml

Repo name and CML controller URL can be configured at the top of this script.
They currently are:

GH Repository name: https://github.com/$REPO
Local CML2 address: $CML

EOT
}

function get_status() {
  if ! tmux list-sessions -F "#S" | grep -qs ^NGROK; then
    echo -n "no "
  fi
  echo "session exists"
}

function remove_secrets() {
  gh api -XDELETE /repos/$REPO/actions/secrets/NGROK_URL
  gh api -XDELETE /repos/$REPO/actions/secrets/USERNAME
  gh api -XDELETE /repos/$REPO/actions/secrets/PASSWORD
  gh api -XDELETE /repos/$REPO/actions/secrets/TOKEN
}

function open() {
  status=$(get_status)
  if [ "$status" = "session exists" ]; then
    tmux attach -t NGROK
  fi
}

function stop() {
  status=$(get_status)
  if [ "$status" = "session exists" ]; then
    tmux kill-session -t NGROK
    remove_secrets
  else
    echo $status
  fi
}

function start() {
  # check if ngrok is running
  if ! curl >/dev/null -sf localhost:4040/api; then
    echo "starting tmux and ngrok"
    tmux &>/dev/null kill-session -t NGROK
    tmux new-session -d -s NGROK
    tmux new-window -t NGROK -n "ngrok" ngrok start --none
    sleep 1
    if ! >/dev/null curl -sf localhost:4040/api; then
      echo "can't start ngrok, failing"
      exit 1
    else
      echo "tmux and ngrok started"
    fi
  fi

  # get the tunnel from the agent and start it, if no tunnel
  TUNNEL=$(curl -sf localhost:4040/api/tunnels | jq -r '.tunnels|map(select(.config.addr == "'$CML'"))[0]|.public_url')
  if [ "$TUNNEL" = "null" ]; then
    DATA='{"proto": "http","addr": "'$CML'","name": "cml"}'
    TUNNEL=$(echo $DATA | curl -sf -XPOST -d@- -H "Content-Type: application/json" localhost:4040/api/tunnels | jq -r '.public_url')
  fi

  # read the public github key for our repo
  read -d' ' GH_KEY_ID GH_KEY <<<"$(gh api /repos/$REPO/actions/secrets/public-key | jq -r '.|.key_id, .key')"

  # make them visible to the ghsecret tool
  export GH_KEY GH_KEY_ID TUNNEL

  # create/update the needed secrets on Github
  # note that gh has a "secret" subcommand... so, the use of ghsecret is
  # only for historical reasons :)
  # https://github.com/cli/cli/releases/tag/v1.4.0
  ghsecret TUNNEL | gh api -XPUT /repos/$REPO/actions/secrets/NGROK_URL --input -
  ghsecret TF_VAR_username | gh api -XPUT /repos/$REPO/actions/secrets/USERNAME --input -
  ghsecret TF_VAR_password | gh api -XPUT /repos/$REPO/actions/secrets/PASSWORD --input -
  ghsecret TF_VAR_token | gh api -XPUT /repos/$REPO/actions/secrets/TOKEN --input -
}

# check if we have everything...
if ! which &>/dev/null ngrok jq gh curl tmux ghsecret; then
  # color="\033[31;40m"
  color="\033[31m"
  nocolor="\033[0m"
  echo
  echo -e $color"Required command is missing!"$nocolor
  echo
  help
  exit 1
fi
if [ "$1" == "start" ]; then
  start
elif [ "$1" == "stop" ]; then
  stop
elif [ "$1" == "open" ]; then
  open
elif [ "$1" == "force" ]; then
  remove_secrets
elif [[ "$1" =~ -h|--help|help ]]; then
  help
elif [ -z "$1" -o "$1" = "status" ]; then
  get_status
else
  help
fi
exit 0
