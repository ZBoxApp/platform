#!/usr/bin/env bash
echo "Login in to docker registry"
docker login -e $DOCKER_EMAIL -u $DOCKER_USER -p $DOCKER_PASS

make zbox-prod

echo "Pushing docker to registry"
docker push zboxapp/zboxnow