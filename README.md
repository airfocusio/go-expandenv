# go-expandenv

Simple helper library to expand environment variables into `interface{}`. The primary use case it to use it to properly expand environment variables into YAML files while properly supporting multi-line environment variables as well:

```yaml
standard: ${ENV_1}
as-number: ${ENV_2:number}
as-boolean: ${ENV_3:boolean}
with-fallback: ${ENV_4:-standard}
```
