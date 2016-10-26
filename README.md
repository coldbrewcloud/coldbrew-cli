# coldbrew-cli

[![Build Status](https://travis-ci.org/coldbrewcloud/coldbrew-cli.svg?branch=master)](https://travis-ci.org/coldbrewcloud/coldbrew-cli)

**coldbrew-cli** automates your Docker container deployments on AWS.

## Getting Started

### Install and Configure CLI

- [Download](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Downloads) CLI executable (`coldbrew` or `coldbrew.exe`) and put it in your $PATH.
- Configure AWS credentials, region, and VPC through [environment variables](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Environment-Variables) or [CLI Flags](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Global-Flags).
- [Prerequisites] Make sure you have [Dockerfile](https://docs.docker.com/engine/reference/builder/) for your application, and, [docker](https://docs.docker.com/engine/installation/) installed in your system.

### Core Concepts

**coldbrew-cli** operates on two simple concepts: applications _(apps)_ and clusters. 

- An **app** is the minimum deployment unit.
- One or more apps can run in a **cluster**, and, they share the computing resources.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/concept.png?v=1" width="350">

This is what a typical deployment workflow might look like:

1. Create new cluster _(See: [cluster-create](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-create))_
2. Create app configuration _(See: [init](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-init))_
3. Development iteration:
  - Make code/configuration changes
  - Deploy app to cluster _(See [deploy](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-deploy))_
  - Check app/cluster status _(See: [status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-status) and [cluster-status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-status))_ and adjust cluster capacity as needed _(See: [cluster-scale](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-scale))_
4. Delete app and its resources _(See: [delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete) )_
5. Delete cluster and its resources _(See: [cluster-delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-delete))_

See [Concepts](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Concepts) for more details.

#### Create Cluster

```bash
coldbrew cluster-create {cluster-name}
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-create.gif?v=1" width="800">

#### Configure App

```bash
coldbrew init --default
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-init-default.gif?v=1" width="800">

#### Deploy App

```bash
coldbrew deploy
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-deploy.gif?v=1" width="800">

#### Check Status

```bash
coldbrew status
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-status.gif?v=1" width="800">

```bash
coldbrew cluster-status
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-status.gif?v=1" width="800">

#### Delete App

```bash
coldbrew delete
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-delete.gif?v=1" width="800">


#### Delete Cluster

```bash
coldbrew cluster-delete
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-delete.gif?v=1" width="800">





