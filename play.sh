#!/usr/bin/env bash
OPTIONS=""
if [[ ! -z "$TF_VAR_username" ]]; then
   OPTIONS="$OPTIONS --env TF_VAR_username=$TF_VAR_username"
fi
if [[ ! -z "$TF_VAR_password" ]]; then
   OPTIONS="$OPTIONS --env TF_VAR_password=$TF_VAR_password"
fi
if [[ ! -z "$TF_VAR_address" ]]; then
   OPTIONS="$OPTIONS --env TF_VAR_address=$TF_VAR_address"
fi

#OPTIONS="$OPTIONS --env ANSIBLE_ROLES_PATH=/ansible/roles"

# docker run -it --rm -v $PWD:/ansible --env PWD="/ansible" --env USER="$USER" $OPTIONS sdwan-test ansible-playbook "$@"
docker run -it --rm -v $PWD:/ansible $OPTIONS sdwan-test ansible-playbook "$@"
