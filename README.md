# Swagger GO

This tiny commandline tool allows you to publish to SwaggerHub using environment
variables or parameters with the SwaggerHub credentials and API path.

The purpose of this application is to simplify the process of publishing new
definitions from deploy pipelines like Buldkite, CircleCI, Github Actions,
Jenkins, etc.

## How to use

From your pipeline you just need to have the [latest release](https://github.com/mijailr/swaggergo/releases) on the `$PATH` of the
runner or agent.

### Simple usage:

```shell script
swaggergo path/to/openapi.yml --type yml --oas 3.0.0 --api mijailr/sample-api --access-token [...]
```

### With environment variables:

```shell script
export SWAGGERHUB_ACCESS_TOKEN="..."
export SWAGGERHUB_API="..."
swaggergo --file path/to/openapi.yml --type yml
```

## Thanks to
This tiny command line tool is inspired on [github-release](https://github.com/buildkite/github-release) from Buildkite.
