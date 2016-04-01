# Copyright (c) 2015 ZBox, Spa. All Rights Reserved.
# See License.txt for license information.
FROM alpine:3.3

RUN apk add --update bash ca-certificates && rm -rf /var/cache/apk/*

RUN mkdir -p /zboxnow
WORKDIR /zboxnow

# Copy over files
ADD ./dist/zboxnow/ /zboxnow/

ADD /docker/dev/config_docker.json /zboxnow/config/config.json
ADD /docker/dev/docker-entry.sh /

RUN chmod +x /docker-entry.sh
ENTRYPOINT /docker-entry.sh

# Create default storage directory
RUN mkdir -p /zboxnow/data/

# Ports
EXPOSE 80
