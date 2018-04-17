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
