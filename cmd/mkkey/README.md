# README.md

This tool creates libsodium compatible encrypted secrets for use with the Github
CLI.  It takes the repo public key that can be obtained via

    /repos/{owner}/{repo}/actions/secrets/public-key

as an environment variable and a secret string as the command line argument.
This produces a base64 encoded encrypted version of the secret string.
