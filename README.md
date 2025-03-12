## What is Hiss? 🐍

The utility package that helps [spf13/viper](https://github.com/spf13/viper) with multiple config sources.

* Read configurations from multiple sources and merge them.
* Read remote resources of AWS (ParameterStore and SecretsManager).
* Get|Put a item from|to DynamoDB tables.

## Config Source URI

* If the value is [ARN](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html), it is treated as an AWS resource. Others are treated as a file path.
* Extensions of file path and ARN resource are used as the content type.
* (viper v1.19) Supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop", "hcl", "tfvars", "dotenv", "env" and "ini".
* If there is no extension, it is regarded as "json".

```sh
--configs="./config.yaml,./config-stg.json"
```

```yaml
configs:
  - "./fixtures/default.yaml"
  - "./fixtures/custom.yaml"
  - "arn:aws:secretsmanager:::secret:stg/apikeys.yaml"
```

## Example
* See [hiss_test.go](./test/hiss_test.go)

## Recommendation

* Do not use global viper instance (https://github.com/spf13/viper?tab=readme-ov-file#viper-or-vipers)