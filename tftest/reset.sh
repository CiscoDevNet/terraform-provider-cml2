#!/bin/bash

rm -f .terraform.lock.hcl
rm -f "terraform.state?(.backup)"
terraform init
terraform plan

