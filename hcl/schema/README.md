# Schema

## Inspect

Reads database schema from database at `<DSN>` and writes to `./schema.hcl`.

```bash
bazel run //hcl/schema/<pkg>:inspect --action_env="DSN=<DSN>"
```

## Apply

Applies `schema.hcl` to database at `<DSN>`.

```bash
DSN="<DSN>" bazel run //hcl/schema/<pkg>:apply
```
