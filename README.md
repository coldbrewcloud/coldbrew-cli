# coldbrew-cli

[![Build Status](https://travis-ci.org/coldbrewcloud/coldbrew-cli.svg?branch=master)](https://travis-ci.org/coldbrewcloud/coldbrew-cli)

tl;dr **coldbrew-cli** automates your Docker container deployment on AWS.

### Objectives

**coldbrew-cli** can provide

* faster access to ECS _(jumpstart with little knowledge on AWS specifics)_
* lower maintenance costs _(most cases you don't even need AWS console or SDK)_
* lessen mistakes by removing boring repetitions
* easier integration with CI

## Getting Started

### Install and Configure CLI

- [Download](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Downloads) CLI executable (`coldbrew` or `coldbrew.exe`) and put it in your `$PATH`.
- Configure AWS credentials, region, and VPC through [environment variables](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Environment-Variables) or [CLI Flags](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Global-Flags).
- Make sure you have [docker](https://docs.docker.com/engine/installation/) installed in your system. You will also need [Dockerfile](https://docs.docker.com/engine/reference/builder/) for your application if you want to build Docker image using **coldbrew-cli**.

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

### Tutorials

Check out tutorials:
- [Running a Node.JS application on AWS](https://github.com/coldbrewcloud/tutorial-nodejs)
- [Running a Slack bot on AWS](https://github.com/coldbrewcloud/tutorial-echo-slack-bot)
- [Running a Meteor application on AWS](https://github.com/coldbrewcloud/tutorial-meteor)
- [Running a Go application on AWS](https://github.com/coldbrewcloud/tutorial-echo)
- [Running a scalable WordPress website on AWS](https://github.com/coldbrewcloud/tutorial-wordpress)

## Core Functions

### Create Cluster

To start deploying your applications, you need to have at least one cluster set up.

```bash
coldbrew cluster-create {cluster-name}
```

[cluster-create](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-create) command will look into your current AWS environment, and, will perform all necessary changes to build the cluster. Note that it can take several minutes until all Docker hosts (EC2 instances) become fully available in your cluster. Use [cluster-status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-status) command to check the status. You can also adjust the cluster's computing capacity using [cluster-scale](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-scale) command.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-create.gif?v=1" width="800">

### Configure App

The next step is prepare the app [configuration file](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-File).

```bash
coldbrew init --default
```

You can manually create/edit your configuration file, or, you can use [init](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-init) command to generate a proper default configuraiton.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-init-default.gif?v=1" width="800">

### Deploy App

Once the configuration file is ready, now you can deploy your app in the cluster.

```bash
coldbrew deploy
```

Basically [deploy](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-deploy) command does:
- build Docker image using your `Dockerfile` _(but this is completely optional if provide your own local Docker image; see [--docker-image](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-deploy#--docker-image) flag)_
- push Docker image to a remote repository (ECR)
- analyze the current AWS environment and setup, and, perform all necessary changes to initiate ECS deployments

Then, within a couple minutes _(mostly less than a minute)_, you will see your new application units up and running.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-deploy.gif?v=1" width="800">

### Check Status

You can use [status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-status) and [cluster-status](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-status) commands to check the running status of your app and cluster respectively.

```bash
coldbrew status
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-status.gif?v=1" width="800">

```bash
coldbrew cluster-status {cluster-name}
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-status.gif?v=1" width="800">

### Delete App

When you no longer need your app, you can remove your app from the cluster using [delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete) command.

```bash
coldbrew delete
```

[delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete) command gathers a list of AWS resources that need to be deleted, and, if you confirm, it will start cleaning them up. It can take several minutes for the full process.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-delete.gif?v=1" width="800">


### Delete Cluster

You can use a cluster for more than one apps, but, when you no longer need the cluster, you use [cluster-delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-delete) command to clean up all the resources.

```bash
coldbrew cluster-delete
```

Similar to [delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete) command, [cluster-delete](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-delete) will delete all AWS resources that are no longer needed. It can take several minutes for the full process.

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-delete.gif?v=1" width="800">

## Documentations

- [Documentations Home](https://github.com/coldbrewcloud/coldbrew-cli/wiki)
- [Managed AWS Resources](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Managed-AWS-Resources)
- [FAQ](https://github.com/coldbrewcloud/coldbrew-cli/wiki/FAQ)
