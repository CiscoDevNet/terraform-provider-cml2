# README.md

This tool creates libsodium compatible encrypted secrets for use with the Github
CLI.  It takes the repo public key information that can be obtained via

    /repos/{owner}/{repo}/actions/secrets/public-key

which looks like

    {
        "key_id": "012345678912345678",
        "key": "2Sg8iYjAxxmI2LvUXpJjkYrMxURPc8r+dB7TJyvv1234"
    }

and a secret value provided via environment variable and creates the proper JSON
to be used with the Github CLI tool `gh`.

The required data looks like this:

    {
        "key_id":"012345678912345678"
        "encrypted_value":"base64encoded_secret",
    }

The tool prints the JSON object that can then be fed into `gh` like this:

    ghsecret ENVVARNAME | gh api -XPUT /some/endpoint --input -

`ENVVARNAME` is the name of the environment variable that holds the secret
string.  In addition, the result from the "public-key" API call must to be
provided as environment variables `GH_KEY` and `GH_KEY_ID`, respectively.

The names of these variables can be changed via command line arguments, if needed.

Here's a silly example:

    $ GH_KEY="$(echo 'qwe' | base64)" GH_KEY_ID="123" ghsecret HOME
    {"key_id":"123","encrypted_value":"bnwu9dXlXcFGYatcXsdpHR0MiiAE3115Mz6wkDrdNACQZSo+1JgPHrhaJCEEnbVpGF5YJMa3tJGGyeb2vqY="}

See the `tunnel.sh` script for a usage example with the 'secrets' API.
