# Prod

## Deploy

```bash
bazel run //prod/<pkg>:push --config=prod --action_env="DOTENV_KEY=<DOTENV_KEY>"
fly deploy --config=prod/<pkg>/fly.toml
```
