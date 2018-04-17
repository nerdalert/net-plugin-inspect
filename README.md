# Inspect Docker Networking Plugin

## Introduction

This binary is intended for testing basic functionality of a 3rd party Docker network driver. The framework is forked from the similar networking tests.


  ```
inspectDockerNetworkingPlugin <insert_driver_name>
  ```


## Inspection and tests

The following inspection steps and tests are performed:

1. The Docker Networking Plugin image is inspected and displayed.

1. The Docker Networking Plugin will be installed if it is not already installed.

1. The Docker Networking Plugin will be uninstalled if it is already installed.

1. A networking is created using the specified plugin.

1. A container is created and attached to the test network created using the 3rd party networking driver.

1. The container and network are deleted to verify the deletion support of the plugin.

1. The 3rd party Docker Network Plugin is removed leaving the host as it was prior to the test.

## Build Instructions

Build the binary with the following:

`go build ./inspectDockerNetworkingPlugin.go`

## Setup

1. Docker Registry credentials need to be specified.

      * You can define them in environment variables.

          * Linux or MacOS

              ```bash
                export DOCKER_USER="my_docker_registry_user_account"
                export DOCKER_PASSWORD="my_docker_registry_user_account_password"
              ```

      * Or you can specify them as arguments to the **inspectDockerNetworkingPlugin** command.

        * **--docker-user**
        * **--docker-password**

      * Otherwise the **inspectDockerNetworkingPlugin** command will prompt for them.

1. By default the **inspectDockerNetworkingPlugin** command uses the following 2 endpoints for communicating to the Docker Hub Registry.

      * Registry Authentication Endpoint: **https://auth.docker.io**
      * Registry API Endpoint: **https://registry-1.docker.io**

    There are 2 ways to override those endpoints:

    * By setting the 2 environment variables below:

      * Linux or MacOS

        ```bash
        export DOCKER_REGISTRY_AUTH_ENDPOINT="https://my_docker_registry_authentication_endpoint"
        export DOCKER_REGISTRY_API_ENDPOINT="https://my_docker_registry_api_enpoint"
        ```

    * Or you can specify them as arguments on the **inspectDockerNetworkingPlugin** command.

        * **--docker-registry-auth-endpoint**
        * **--docker-registry-api-endpoint**

## Syntax

```
Inspects a Docker Networking Plugin to see if it conforms to best practices.

Syntax: inspectDockerNetworkingPlugin [options] dockerNetworkingPlugin

Options:
  -docker-user string
    	 Docker User ID.  This overrides the DOCKER_USER environment variable.
  -docker-password string
    	 Docker Password.  This overrides the DOCKER_PASSWORD environment variable.
  -docker-registry-api-endpoint string
    	 Docker Registry API Endpoint. This overrides the DOCKER_REGISTRY_API_ENDPOINT environment variable. (default "https://registry-1.docker.io")
  -docker-registry-auth-endpoint string
    	 Docker Registry Authentication Endpoint. This overrides the DOCKER_REGISTRY_AUTH_ENDPOINT environment variable. (default "https://auth.docker.io")
  -help
    	 Help on the command.
  -html
    	 Generate HTML output.
  -json
    	 Generate JSON output.
  -verbose
    	 Displays more verbose output.

  dockerNetworkingPlugin
	The Docker Networking Plugin to inspect. This argument is required.
```

## Output

The **inspectDockerNetworkingPlugin** command can generate 3 types of output results:

1. Messages sent to stdout (Default)
1. HTML local file
1. JSON sent to stdout

By default the **inspectDockerNetworkingPlugin** command generates messages sent to stdout. You can specify the **--json** option which overrides and replaces the messages sent to stdout.
You can also specify the **--html** option which generates an HTML report. And both **--json** and **--html** can be specified at the same time.

#### Default Output:

The following command produces the default output results:

`./inspectDockerNetworkingPlugin weaveworks/net-plugin:latest_release`

Results:

```
**************************************************************************************************************************************************************************************************
* Docker Networking Plugin: weaveworks/net-plugin:latest_release
**************************************************************************************************************************************************************************************************

**************************************************************************************************************************************************************************************************
* Step #1 Inspecting the Docker Networking Plugin: weaveworks/net-plugin:latest_release ...
**************************************************************************************************************************************************************************************************
Passed:   Docker Networking Plugin image weaveworks/net-plugin:latest_release has been inspected.

**************************************************************************************************************************************************************************************************
* Step #2 Docker Networking Plugin information
**************************************************************************************************************************************************************************************************
+-------------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| Docker Networking Plugin: | weaveworks/net-plugin:latest_release                                                                                                                                 |
| Description:            | Weave Net plugin for Docker                                                                                                                                          |
| Documentation:          | https://weave.works                                                                                                                                                  |
| Digest:                 | sha256:5016b9e5596df4c700546421ffea86a8a6d9e8c65cdf5306afb14adeedee102a                                                                                              |
| Base layer digest:      | sha256:34165d7d46d9f543a80f82774913949a6808149a0a7192c1e98aebf98081fd3d                                                                                              |
| Docker version:         | 17.05.0-ce                                                                                                                                                           |
| Interface Socket:       | weave.sock                                                                                                                                                           |
| Interface Socket Types: | docker.networkdriver/1.0                                                                                                                                             |
| IpcHost:                | false                                                                                                                                                                |
| PidHost:                | false                                                                                                                                                                |
| Entrypoint:             | /home/weave/launch.sh                                                                                                                                                |
| WorkDir:                |                                                                                                                                                                      |
| User:                   |                                                                                                                                                                      |
+-------------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------+

**************************************************************************************************************************************************************************************************
* Step #3 Installing the Docker Networking plugin weaveworks/net-plugin:latest_release ...
**************************************************************************************************************************************************************************************************
Passed:   Docker networking plugin weaveworks/net-plugin:latest_release has been installed successfully.

**************************************************************************************************************************************************************************************************
* Step #4 Testing the Docker network creation using plugin: weaveworks/net-plugin:latest_release ...
**************************************************************************************************************************************************************************************************
Passed:   Docker network was created using plugin weaveworks/net-plugin:latest_release

**************************************************************************************************************************************************************************************************
* Step #5 Testing the Docker network deletion using plugin: weaveworks/net-plugin:latest_release ...
**************************************************************************************************************************************************************************************************
Passed:   Docker network was removed using plugin weaveworks/net-plugin:latest_release

**************************************************************************************************************************************************************************************************
* Step #6 Removing the Docker networking plugin
**************************************************************************************************************************************************************************************************
Passed:   Docker network plugin weaveworks/net-plugin:latest_release was removed.

**************************************************************************************************************************************************************************************************
* Summary of the inspection for the Docker Networking Plugin: weaveworks/net-plugin:latest_release
**************************************************************************************************************************************************************************************************

Report Date: Mon Apr 16 13:16:52 2018
Operating System: Operating System: MacOS darwin Version: 10.12.6
Architecture: amd64
Docker version 18.02.0-ce, build fc4de44


Passed:   Docker Networking Plugin image weaveworks/net-plugin:latest_release has been inspected.
Passed:   Docker networking plugin weaveworks/net-plugin:latest_release has been installed successfully.
Passed:   Docker network was created using plugin weaveworks/net-plugin:latest_release
Passed:   Docker network was removed using plugin weaveworks/net-plugin:latest_release
Passed:   Docker network plugin weaveworks/net-plugin:latest_release was removed.

The inspection of the Docker networking plugin weaveworks/net-plugin:latest_release has completed.
```


### Inspect a Docker Networking Plugin with JSON Output


```
$> [inspect_docker_networking_plugin]$ ./inspectDockerNetworkingPlugin --json weaveworks/net-plugin:latest_release | jq
```

Note: The output was piped to the **jq** command to display it "nicely".

#### Output:

```
{
  "Date": "Mon Apr 16 13:14:14 2018",
  "SystemOperatingSystem": "Operating System: MacOS darwin Version: 10.12.6",
  "SystemArchitecture": "amd64",
  "SystemDockerVersion": "Docker version 18.02.0-ce, build fc4de44",
  "DockerLogginPlugin": "weaveworks/net-plugin:latest_release",
  "Description": "Weave Net plugin for Docker",
  "Documentation": "https://weave.works",
  "DockerNetworkingPluginDigest": "sha256:5016b9e5596df4c700546421ffea86a8a6d9e8c65cdf5306afb14adeedee102a",
  "BaseLayerImageDigest": "sha256:34165d7d46d9f543a80f82774913949a6808149a0a7192c1e98aebf98081fd3d",
  "DockerVersion": "17.05.0-ce",
  "Entrypoint": "/home/weave/launch.sh",
  "InterfaceSocket": "weave.sock",
  "InterfaceSocketTypes": "docker.networkdriver/1.0",
  "WorkDir": "",
  "User": "",
  "IpcHost": false,
  "PidHost": false,
  "Errors": 0,
  "Warnings": 0,
  "HTMLReportFile": "",
  "VulnerabilitiesScanURL": "",
  "Results": [
    {
      "Status": "Passed",
      "Message": "Docker Networking Plugin image weaveworks/net-plugin:latest_release has been inspected."
    },
    {
      "Status": "Passed",
      "Message": "Docker networking plugin weaveworks/net-plugin:latest_release has been installed successfully."
    },
    {
      "Status": "Passed",
      "Message": "Docker network was created using plugin weaveworks/net-plugin:latest_release"
    },
    {
      "Status": "Passed",
      "Message": "Docker network was removed using plugin weaveworks/net-plugin:latest_release"
    },
    {
      "Status": "Passed",
      "Message": "Docker network plugin weaveworks/net-plugin:latest_release was removed."
    }
  ]
}
```

<a name="inspect-networking-plugin-html">

### Inspect a Docker Networking Plugin with HTML Output

#### To inspect the  Docker Networking Plugin "weaveworks/net-plugin:latest_release" with HTML Output:

```
$> ./inspectDockerNetworkingPlugin --html weaveworks/net-plugin:latest_release
```

![HTML Output Image](/screenshots/screenshots-netplugin.png "HTML Output")
