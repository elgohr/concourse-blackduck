workflow "Publish Docker" {
  resolves = [
    "logout",
  ]
  on = "push"
}

action "login" {
  uses = "actions/docker/login@master"
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "publish" {
  uses = "elgohr/Publish-Docker-Github-Action@master"
  args = "lgohr/blackduck-resource"
  needs = ["login"]
}

action "logout" {
  uses = "actions/docker/cli@master"
  args = "logout"
  needs = ["publish"]
}
