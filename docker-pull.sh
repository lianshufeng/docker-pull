#!/bin/bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:/work lianshufeng/docker-pull hello-world