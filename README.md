# Concourse Blackduck Resource
This is a [Concourse](https://concourse-ci.org/) resource for [Blackduck](https://www.blackducksoftware.com).  

__The State of this resource is early alpha, so please take care and give feedback. Thank you!__

## Installing

`Shortcut`: [Pipeline example](https://github.com/elgohr/concourse-blackduck/blob/master/example-pipeline.yml)

Use this resource by adding the following to
the `resource_types` section of a pipeline config:

```yaml
resource_types:
- name: blackduck
  type: docker-image
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

* `url`: *Required* URL of your Blackduck instance e.g. `https://my-synopsys.com/blackduck`.
* `username`: *Optional* Username, which is used to authenticate on Blackduck.
* `password`: *Optional* Password, which is used to authenticate on Blackduck.
* `token`: *Optional* API-Token, which is used to authenticate on Blackduck.
* `name`: *Required* Project name in Blackduck.

## `in`: Nothing yet

## `out`: Analysis
The resource will analyse your provided content and push it to the provided Blackduck instance.

### Parameters

```yaml
  - put: my-blackduck
    params: {directory: source-code}
```

* `directory`: *Required.* The path of the repository to analyze.

