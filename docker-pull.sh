#!/bin/bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${pwd}:/work lianshufeng/docker-pull -i hello-world