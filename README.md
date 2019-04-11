# Concourse Blackduck Resource
This is a [Concourse](https://concourse-ci.org/) resource for [Blackduck](https://www.blackducksoftware.com).  

__The State of this resource is early alpha, so please take care and give feedback. Thank you!__

## Installing

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
```

* `url`: *Required.* URL of your Blackduck instance e.g. `https://my-synopsys.com/blackduck`.
* `username`: *Required.* Username, which is used to authenticate on Blackduck.
* `password`: *Required.* Password, which is used to authenticate on Blackduck.

## `in`: Nothing yet

## `out`: Analysis
The resource will analyse your provided content and push it to the provided Blackduck instance.
