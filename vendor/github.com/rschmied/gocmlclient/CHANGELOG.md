# gocmlclient Changelog

Lists the changes in the gocmlclient package.

## Version 0.0.16

- added new API endpoints for groups and users
- tried to apply some consistency to func names
- add 201 return code to the list of "OK" codes
- allow to set "groups" when creating/updating a lab
- updated dependencies

## Version 0.0.15

- made node configuration a pointer to differentiate between
  "no configuration" (null), "empty configuration" and "specific
  configuration". With a null configuration, the default configuration
  from the node definition will be inserted if there is one
- added Version var/func, moved NewClient() to New()
- bump go to 1.19 and vendor deps

## Version 0.0.12

- Realized that the empty tags removal from 0.0.11 caused a regression.
  node tags are always returned/set even when there's no tags... in that
  case, the empty list is returned or needs to be provided. See 0.0.3 comment.
- Test coverage improvement

## Version 0.0.8 to 0.0.11

- Added most of the doc string for exported functions.
- reversed the sorting of images for the image definitions.
- sort image definitions by their ID. Lists have the newest (highest version)
  image as the first element.
- updated dependencies.
- have InterfaceCreate accept a slot value (not a pointer). A negative slot
  indicates "don't specify a slot", this was previously indicated by nil.
- added more values to the ImageDefinition and Nodedefinition structs.
- added a link unit test.
- more node attributes can be updated when a node is DEFINED_ON_CORE
- NodeCreate removes a node now when the 2nd API call fails. The 2nd call is
  needed to update certain attributes which are not accepted in the actual
  create API (POST).
- move the upper version for the version constraint from <2.6.0 to <3.0.0.
- omit empty tags on update.

## Version 0.0.5 to 0.0.7

- refactored the code so that interfaces are read in one go ("data=true"). This
  without this, only a list of interface IDs is returned by the API. With this,
  the API returns a list of complete interface object.
- Implement the same approach for nodes (0.0.6).
- updated dependencies.
- Due to the data=true option, restrict the code to only work with 2.4.0 and later.

## Version 0.0.4

- added additional locking to prevent races
- improved test coverage, esp. w/ cache usage

## Version 0.0.3

- Fixed node tag list update. To delete all tags from a node, an empty tag list
  must be serialized in the `PATCH` JSON.  This was prevented by having
  `omitempty` in the struct.  Fixed  
- Also moved the `ctest` cmd fro the terraform provider repo to the code base.

## Versions prior to 0.0.3

Nothing in particular to be noteworthy -- just huge chunks of initial code
refactoring.
