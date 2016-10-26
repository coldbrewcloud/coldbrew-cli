# coldbrew-cli

[![Build Status](https://travis-ci.org/coldbrewcloud/coldbrew-cli.svg?branch=master)](https://travis-ci.org/coldbrewcloud/coldbrew-cli)

**coldbrew-cli** automates your Docker container deployments on AWS.

## Getting Started

### Install and Configure CLI

- [Download](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Downloads) CLI executable (`coldbrew` or `coldbrew.exe`) and put it in your $PATH.
- Configure AWS credentials, region, and VPC through [environment variables](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Environment-Variables) or [CLI Flags](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Global-Flags).
- [Prerequisites] Make sure you have [Dockerfile](https://docs.docker.com/engine/reference/builder/) for your application, and, [docker](https://docs.docker.com/engine/installation/) installed in your system.

### Core Concepts

**coldbrew-cli** has 2 simple concepts, applications (apps) and clusters. Long story short, apps are the minimum deployment units, and, clusters are where one or more apps are running together sharing some of AWS resources. Typical setup would be a couple of applications (for your projects) running in a cluster. See [Concepts](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Concepts) for more details.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/concept.png?v=1" width="350">

### Typical Workflow

1. Create new cluster _(See [cluster-create](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-create) command)_
2. Create app configuration _(See [init](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-init) command)_
3. Development iteration:
  - Make code/configuration changes.
  - Deploy app _(See [deploy](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-deploy) command)_
  - Check the status _(See [status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-status) and [cluster-status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-status) commands)_ and adjust cluster capacity as needed _(See [cluster-scale](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-scale) command)_
4. Delete app and its resources _(See [delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete) command)_
5. Delete cluster and its resources _(See [cluster-delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-delete) command)_


### Create Cluster

```bash
coldbrew cluster-create cluster1
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-create.gif?v=1" width="800">

### Init App

Here we assume that you already have your `Dockerfile`

```bash
coldbrew init --default
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-init-default.gif?v=1" width="800">

### Deploy App

```bash
coldbrew deploy
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-deploy.gif?v=1" width="800">





