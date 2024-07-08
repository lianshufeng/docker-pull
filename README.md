# Docker Download Tool


## Features

- **Thread**
- **Cache**
- **Proxy**
- **Mirror**
- **Manifest: v1 and v2** - https://docs.docker.com/registry/

## Basic
```bash
docker-pull [-config] <imageName>

eg : docker-pull -proxy http://127.0.0.1:1080 -thread 5 nginx
  -architecture string
        platform.architecture (default "amd64")
  -buffByte int
        download buffByte (default 10485759)
  -cache string
        cache directory (default "_cache")
  -docker_api_version string
        DOCKER_API_VERSION (default "1.30")
  -load
        load image (default true)
  -m string
        mirror url , docker.chatsbot.org
  -os string
        platform.os (default "linux")
  -proxy string
        proxy server , http://127.0.0.1:1080
  -thread int
        thead number (default 6)
  -variant string
        platform.variant
```

## Usage
- simple
```bash
docker-pull nginx
docker-pull mysql/mysql-server:8.0
docker-pull nginx@sha256:d987469c32fecb3f839a4606671eb2bb308039a3a6b2d086341769da3931b9b6
```
- with proxy
```bash
docker-pull -proxy http://127.0.0.1:1080 nginx
```
- with mirror
```bash
docker-pull -m docker.chatsbot.org nginx
```

## Docker 

- x86
```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:/work lianshufeng/docker-pull nginx
```

- arm
```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:/work lianshufeng/docker-pull -architecture arm nginx
```


## Installation
1. Clone the repository:
```bash
git clone https://github.com/lianshufeng/docker-pull.git
```
2. Build:
```bash
cd core/
go build ./
```