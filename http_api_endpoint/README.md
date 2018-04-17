# http_api_endpoint

## Introduction

The **http_api_endpoint** is an HTTP Server that can be used to test docker logging plugins that do not support the read log api and instead send data to an API Endpoint running on an external server.
The [Sumo Logic Logging Plugin](https://store.docker.com/plugins/sumologic-logging-plugin) is one example.

You can configure those docker logging plugins to send their logging data to the **http_api_endpoint** HTTP Server for testing the plugin and then code a script to retrieve the logs using a curl command.

## Syntax

```
./http_api_endpoint [options]
```

options:

 * **--port**    (The port for the **http_api_endpoint** HTTP Server to listen on. Defaults to port 80)
 * **--debug**   (write debugging information)
 * **--help**    (display the command help)

## Using and testing the **http_api_endpoint** HTTP Server

The **curl** command can be used to test and use the **http_api_endpoint** HTTP Server.

* Issue the following curl command to send new data to the **http_api_endpoint**:

  ```
  # DATA='Hello World!'
  # curl -s -X POST -d "${DATA}" http://127.0.0.1:80
  ```

  Note: if any data was previously sent, it will be replaced.

* Issue the following curl command to send data to the **http_api_endpoint** and append that data to the already collected data:

  ```
  # DATA='Hello World!'
  # curl -s -X POST -d "${DATA}" http://127.0.0.1:80
  ```

* Issue the following curl command to retrieve the data from the http_api_endpoint:

  ```
  # curl -s -X GET http://127.0.0.1:80
  ```
  ```
  Hello World!
  ```

* Issue the following curl command to erase any data currently collected by the http_api_endpoint:

  ```
  # curl -s -X DELETE http://127.0.0.1:80
  ```

* To Terminate:

  ```
  # curl -s http://127.0.0.1:80/EXIT
  ```

## Example of using the http_api_endpoint HTTP Server for the Sumo Logic Logging Plugin

### Script to run a container to test the Sumo Logic Logging Plugin

```
# cat test_sumologic_plugin.sh
```
```
#!/usr/bin/env bash

#######################################################################################################################################
# This bash script tests the Sumo Logic Docker Logging Plugin
#
# The script accepts 1 command line parameter:  Sumo Logic Docker Logging PLugin
#
#######################################################################################################################################
# Gary Forghetti
# Docker Inc.
#######################################################################################################################################

#######################################################################################################################################
# Make sure the Sumo Logic Docker Logging Plugin was specified on the command line
#######################################################################################################################################
DOCKER_LOGGING_PLUGIN=$1
if [[ -z $DOCKER_LOGGING_PLUGIN ]]; then
    printf 'You must specify the Docker Loggin Plugin!\n'
    exit 1  
fi

HTTP_API_ENDPOINT='http://localhost:80'

#######################################################################################################################################
# Check to make sure the http_api_endpoint HTTP Server is running
#######################################################################################################################################
curl -s -X POST "${HTTP_API_ENDPOINT}"
if [[ $? -ne 0 ]]; then
    printf 'Unable to connect to the HTTP API Endpoint: '"${HTTP_API_ENDPOINT}"'!\n'
    exit 1
fi

#######################################################################################################################################
# Run a alpine container with the plugin and send data to it
#######################################################################################################################################
docker container run \
--rm \
--log-driver="${DOCKER_LOGGING_PLUGIN}" \
--log-opt sumo-url="${HTTP_API_ENDPOINT}" \
--log-opt sum-sending-interval=5s \
--log-opt sumo-compress=false \
--volume $(pwd)/quotes.txt:/quotes.txt \
alpine:latest \
sh -c 'cat /quotes.txt;sleep 10'

exit $?
```

### Script to retrieve the logging data from the http_api_endpoint HTTP Server

```
# cat get_sumologic_logs.sh
```
```
#!/usr/bin/env sh

#######################################################################################################################################
# This bash script retrieves any data logged to the http_api_endpoint HTTP Server.
#######################################################################################################################################
# Gary Forghetti
# Docker Inc.
#######################################################################################################################################

curl -s -X GET http://127.0.0.1:80

```

### To test the Sumo Logic Logging Plugin

```
./inspectDockerLoggingPlugin --verbose --html --test-script ./test_sumologic_plugin.sh --get-logs-script ./get_sumologic_logs.sh dockerstorestaging/sumologic-docker-logging-driver:1.0.2
```
