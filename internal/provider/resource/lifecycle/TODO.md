# TODO

When updating the lab and going from DEFINED->STARTED->STOPPED->STARTED,
computed attributes are not properly updated.  See `modify_plan.go`.  The "fix"
in place marks them all as unknown.  Ideally, they are only marked as unknown
for specific state changes.  But it works for now.

```plain
│ Error: Provider produced inconsistent result after apply
│ 
│ When applying changes to cml2_lifecycle.this, provider "provider[\"registry.terraform.io/ciscodevnet/cml2\"]" produced an unexpected new value:
│ .nodes["c3d439a2-b1a6-4d6e-b3a0-b93c55536cda"].cpus: was cty.NumberIntVal(0), but now cty.NumberIntVal(1).
│ 
│ This is a bug in the provider, which should be reported in the provider's own issue tracker.
╵
```
