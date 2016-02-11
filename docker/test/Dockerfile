# Copyright (c) 2015 ZBox, Spa. All Rights Reserved.
# See License.txt for license information.
FROM ubuntu:14.04

# Install Dependancies
RUN apt-get update && apt-get install -y curl
RUN curl -sL https://deb.nodesource.com/setup | bash -


RUN mkdir -p /zboxnow
WORKDIR /zboxnow


# Copy over files
ADD ./dist/zboxnow/ /zboxnow/

ADD /docker/test/config_docker.json /zboxnow/config/config.json
ADD /docker/test/docker-entry.sh /

RUN chmod +x /docker-entry.sh
ENTRYPOINT /docker-entry.sh

# Create default storage directory
RUN mkdir -p /zboxnow/data/

# Ports
EXPOSE 80
