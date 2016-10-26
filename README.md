# coldbrew-cli

[![Build Status](https://travis-ci.org/coldbrewcloud/coldbrew-cli.svg?branch=master)](https://travis-ci.org/coldbrewcloud/coldbrew-cli)

**coldbrew-cli** automates your Docker container deployments on AWS.

## Getting Started

### Install CLI

**coldbrew-cli** is distributed as a binary package.

- Below are the available downloads for the latest version of **coldbrew-cli**. Please download the proper package for your operations system and architecture.
  - [Linux 64-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/linux/amd64/coldbrew) / [32-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/linux/386/coldbrew)
  - [Mac 64-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/darwin/amd64/coldbrew) / [32-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/darwin/386/coldbrew)
  - [Windows 64-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/windows/amd64/coldbrew.exe) / [32-bit](https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/cli/windows/386/coldbrew.exe)
- **coldbrew-cli** is a single binary executable (`coldbrew` or `coldbrew.exe` for Windows). Once downloaded, you can move or copy the executable into your `$PATH` (e.g. `/usr/local/bin` on Mac).

Alternatively you can build the executable from the source:
- You need [Go](https://golang.org/) to build the source.
- `git clone https://github.com/coldbrewcloud/coldbrew-cli.git`
- `cd coldbrew-cli`
- `go build -o coldbrew`
- Now you have `coldbrew` executable.

### Configure CLI

**coldbrew-cli** uses several environment variables to take your AWS keys, region name, and, VPC ID.

- `$AWS_ACCESS_KEY_ID`: AWS Access Key ID _(required)_ 
- `$AWS_SECRET_ACCESS_KEY`: AWS Secret Access Key _(required)_
- `$AWS_REGION`: AWS  region name _(default: `"us-west-2"`)_
- `$AWS_VPC`: AWS VPC ID _(optional if you have [default VPC configured](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html), required otherwise)_

See [CLI Flags](https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Global-Flags) in case you do not want to use environment variables.

### Core Concepts

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/concept.png" width="350">

**coldbrew-cli** has 2 simple concepts, applications (apps) and clusters. Long story short, apps are the minimum deployment units, and, clusters are where one or more apps are running together sharing some of AWS resources. Typical setup would be a couple of applications (for your projects) running in a cluster. See [Concepts](https://github.com/coldbrewcloud/coldbrew-cli/wiki/Concepts) for more details.

### Create Cluster

```bash
coldbrew cluster-create cluster1
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-cluster-create.gif" width="700">

### Init App

Here we assume that you already have your `Dockerfile`

```bash
coldbrew init --default
```

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-init-default.gif" width="700">

### Deploy App

<img src="https://raw.githubusercontent.com/coldbrewcloud/assets/master/coldbrew-cli/command-deploy.gif" width="700">





