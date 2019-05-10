workflow "Publish Docker" {
  resolves = [
    "logout",
  ]
  on = "push"
}

action "login" {
  uses = "actions/docker/login@8cdf801b322af5f369e00d85e9cf3a7122f49108"
  secrets = [
    "DOCKER_USERNAME",
    "DOCKER_PASSWORD",
  ]
}

action "publish" {
  uses = "elgohr/Publish-Docker-Github-Action@1.0"
  args = "lgohr/blackduck-resource"
  needs = ["login"]
}

action "logout" {
  uses = "actions/docker/cli@8cdf801b322af5f369e00d85e9cf3a7122f49108"
  args = "logout"
  needs = ["publish"]
}
