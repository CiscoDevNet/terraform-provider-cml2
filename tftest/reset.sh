#!/bin/bash

# set -x
shopt -s extglob

rm -f .terraform.lock.hcl
if [ -n "$1" ]; then
    rm -f terraform.tfstate?(.backup)
fi

terraform init
terraform plan

