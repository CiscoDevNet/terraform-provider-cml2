#!/bin/bash

version=$(git describe | sed -E 's/^v(.*)$/\1/')

cat <<EOF
{
  "version": "$version"
}
EOF

