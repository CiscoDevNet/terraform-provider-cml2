# TODO

As this provider is "work-in-progress", there's still plenty of work to do.

- additional resources and data sources
- better test coverage(unit/acceptance)
- figure out if/how <https://github.frangipane.io/> fits this provider

Especially the first bullet requires some discussion in terms of what makes
sense and what doesn't.

## Items Done

- added external connector data source
- added group and user resources and data sources
- basic acceptance tests
- documentation (content, consistency)
- added image definition data source
- add CA pem file to provider config
- move the actual golang CML client library to its own repo
- naming consistency (cml vs cml2) -- should be OK by now
- image schema unit test
- fix documentation of data sources ("optional schema props")
- improved documentation and examples

## Not doing

- "filter" implementation for lab details data source (if needed at all)

## References

- Release and Publish a Provider <https://learn.hashicorp.com/tutorials/terraform/provider-release-publish?in=terraform/providers>
- GitHub Encrypted Secrets <https://docs.github.com/en/actions/security-guides/encrypted-secrets>
- GoReleaser GitHub Action <https://github.com/marketplace/actions/goreleaser-action>
