#!/bin/bash

# set -x
shopt -s extglob

rm -f .terraform.lock.hcl
rm -f terraform.tfstate?(.backup)

terraform init
terraform plan

