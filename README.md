# Sensu Stackdriver Handler

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Resource definition](#resource-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The Sensu Stackdriver Handler is a [Sensu Handler][6] that sends Sensu
Go collected metrics to Google Stackdriver. Leverage Sensu Go to
collect and process metrics from a plethora of data sources, from
Nagios service check plugin executions to the Sensu Backend and Agent
data ingestion APIs. Provide the Handler with a Google Stackdriver
project ID and begin storing time series in Stackdriver with Sensu Go.

## Usage examples

Help:

```
Send Sensu Go collected metrics to Google Stackdriver

Usage:
  sensu-stackdriver-handler [flags]
  sensu-stackdriver-handler [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -h, --help                help for sensu-stackdriver-handler
  -p, --project-id string   The Google Cloud Project ID

Use "sensu-stackdriver-handler [command] --help" for more information about a command.
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add portertech/sensu-stackdriver-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/project/sensu-stackdriver-handler].

### Resource definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: stackdriver
  namespace: default
spec:
  command: sensu-stackdriver-handler -p my-project-id-123
  type: pipe
  runtime_assets:
  - sensu-stackdriver-handler
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-stackdriver-handler repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/handler-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/handler-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[7]: https://github.com/sensu-community/handler-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
