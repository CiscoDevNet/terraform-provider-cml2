# TODO

As this provider is "work-in-progress", there's still plenty of work to do:

- documentation (content, consistency)
- tests (none so far)
- "filter" implementation for lab details data source (if needed at all)
- naming consistency (cml vs cml2)?
- additional resources and data sources

Especially the last bullet requires some discussion in terms of what makes
sense and what doesn't.

In addition, I am still somewhat unclear about the schema when it comes to more
complex / nested data structures.  As indicated [here](https://discuss.hashicorp.com/t/question-nested-attribute-lists-result-in-tolist-json-output-why/39200) I am
having a couple of questions what the proper approach is.

There's some documentation around design [here](https://github.com/hashicorp/terraform-plugin-framework/tree/main/docs/design) but that wasn't really conclusive; it also seemed to be referencing
outdated material.

## Items Done

- add CA pem file to provider config

## References

- Release and Publish a Provider <https://learn.hashicorp.com/tutorials/terraform/provider-release-publish?in=terraform/providers>
- GitHub Encrypted Secrets <https://docs.github.com/en/actions/security-guides/encrypted-secrets>
- GoRelease GitHub Action <https://github.com/marketplace/actions/goreleaser-action>
