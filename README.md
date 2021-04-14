# Concourse Blackduck Resource
[![Actions Status](https://github.com/elgohr/concourse-blackduck/workflows/Test/badge.svg)](https://github.com/elgohr/concourse-blackduck/actions)
[![Actions Status](https://github.com/elgohr/concourse-blackduck/workflows/Publish%20Master/badge.svg)](https://github.com/elgohr/concourse-blackduck/actions)


This is a [Concourse](https://concourse-ci.org/) resource for [Blackduck](https://www.blackducksoftware.com).  

## Installing

`Shortcut`: [Pipeline example](https://github.com/elgohr/concourse-blackduck/blob/master/example-pipeline.yml)

Use this resource by adding the following to
the `resource_types` section of a pipeline config:

```yaml
resource_types:
- name: blackduck
  type: registry-image
  source:
    repository: lgohr/blackduck-resource
    tag: latest
```

## Source configuration

Configure as follows:

```yaml
resources:
- name: my-blackduck
  type: blackduck
  source:
    url: https://my.blackduck.server
    username: ((my-secret-username))
    password: ((my-secret-password))
    name: myScanProject
```

| Argument        | Mandatory               | Description                                                                                |
|-----------------|-------------------------|--------------------------------------------------------------------------------------------|
| `url`           | *Mandatory*             | URL of your Blackduck instance e.g. `https://my-synopsys.com/blackduck`.                   |
| `name`          | *Mandatory*             | Project name in Blackduck.                                                                 |
| `username`      | *Mandatory*             | Username, which is used to authenticate on Blackduck.                                      |
| `password`      | *Mandatory*             | Password, which is used to authenticate on Blackduck.                                      |
| `insecure`      | *Optional*              | In case your Blackduck uses a self-signed certificate, it's pinned with the first request. |
| `proxy-host`    | *Optional*              | In case your Concourse needs to use a proxy to connect to Blackduck.                       |
| `proxy-port`    | *Optional*              | In case your Concourse needs to use a proxy to connect to Blackduck.                       |
| `proxy-username`| *Optional*              | In case your Concourse needs to use a proxy to connect to Blackduck.                       |
| `proxy-password`| *Optional*              | In case your Concourse needs to use a proxy to connect to Blackduck.                       |

It seems like Blackduck doesn't support Tokens for API-Access (in the scanner it would work fine).  
As the configuration should be clean and understandable, the token is not supported. Sorry.

## `in`: Get Results
The resource will provide the latest version changes on Blackduck as a file for later use.

## `out`: Analysis
The resource will analyse your provided content and push it to the provided Blackduck instance.

### Parameters

```yaml
  - put: my-blackduck
    params: {directory: source-code}
```

* `directory`: *Required.* The path of the repository to analyze.
