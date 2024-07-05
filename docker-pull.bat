@echo off
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v %CD%:/work lianshufeng/docker-pull -i hello-world