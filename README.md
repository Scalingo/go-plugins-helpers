# go-plugins-helpers v1.3.0

A collection of helper packages to extend Docker Engine in Go

 Plugin type   | Documentation | Description
 --------------|---------------|--------------------------------------------------
 Authorization | [Link](https://docs.docker.com/engine/extend/authorization/)   | Extend API authorization mechanism
 Network       | [Link](https://docs.docker.com/engine/extend/plugins_network/) | Extend network management
 Volume        | [Link](https://docs.docker.com/engine/extend/plugins_volume/)  | Extend persistent storage
 IPAM          | [Link](https://github.com/docker/libnetwork/blob/master/docs/ipam.md) | Extend IP address management
 Graph (experimental) | [Link](https://github.com/docker/cli/blob/master/docs/extend/plugins_graphdriver.md)| Extend image and container fs storage

See the [understand Docker plugins documentation section](https://docs.docker.com/engine/extend/plugins/).

## Test Environment

In a non-Docker environment, you may want to define the environment variable `PLUGIN_SPEC_DIR` to a user-writable folder such as:

```shell
PLUGIN_SPEC_DIR=$(pwd)/_dev go test -v ./...
```

## Release a New Version

Bump new version number in:

- `CHANGELOG.md`
- `README.md`

Commit, tag and create a new release:

```sh
git add CHANGELOG.md README.md
git commit -m "Bump v1.3.0"
git tag v1.3.0
git push origin master
git push --tags
hub release create v1.3.0
```

The title of the release should be the version number and the text of the release is the same as the changelog.
