# Flipadelphia

*Flipadelphia flips your features*

<img src="http://i.imgur.com/28TTvje.gif" alt="flipadelphia"/>

## Installation
* Flipadelphia uses a config file to keep track of settings for runtime environments. By default, the file is expected
    to be in the ~/.flipadelphia directory. You can also manually set the config with the -c/--config flag. An example
    of the config file is in the config subpackage.
* To handle dependencies, Flipadelphia uses the [glide](https://github.com/Masterminds/glide) package manager. Glide
    installs dependencies into the ```vendor``` directory within a project, and handles versioning of dependencies.
    If you don't want to use glide, you can always just manually add the dependencies, but you risk running into a
    version mismatch since glide locks dependencies down to specific commits.

```sh
$ mkdir ~/.flipadelphia
$ cp config/config.example.json ~/.flipadelphia/config.json
$ glide install
```

## Building

```sh
$ make build
```

## Running
```sh
$ ./flipadelphia -h
NAME:
   flipadelphia - flipadelphia flips your features

USAGE:
   flipadelphia [global options] command [command options] [arguments...]

VERSION:
   dev-build

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --env, -e "development"					                An environment from the config.json file to use [$FLIPADELPHIA_ENV]
   --config "/Users/samfonseca/.flipadelphia/config.json"	Path to the config file. [$FLIPADELPHIA_CONFIG]
   --help, -h							                    show help
   --version, -v						                    print the version
```

## Usage

### Setting a feature

```sh
$ curl -s localhost:3006/features/feature1?scope=user-1 | jq .
{
  "name": "feature1",
  "value": "on",
  "data": "true"
}
```

### Checking a feature

If the feature has been set on the scope, ```data``` is ```true```.
```sh
$ curl -s localhost:3006/features/feature1?scope=user-1 | jq .
{
  "name": "feature1",
  "value": "on",
  "data": "true"
}
```

If the feature has never been set on the scope, ```data``` is ```false```.
```sh
$ curl -s localhost:3006/features/unset_feature/?scope=user-1 | jq .
{
  "name": "unset_feature",
  "value": "",
  "data": "false"
}
```

## Performance

* Flipadelphia uses BoltDB as the persistence layer. BoltDB fits the nature of a feature flipping service because it's a read-optimized database and features are typically checked far more often than set.
* To keep the response time as low as possible, all the check endpoints come without authorization. The thinking here is that theres no harm in someone checking a feature. If thats an issue, use a uuid for the feature and scope and keep a mapping of those externally. Authentication will be added to the set endpoint soon.

```sh
ab -n 1000 -c 20 localhost:3006/features/feature1?scope=user-1
This is ApacheBench, Version 2.3 <$Revision: 1663405 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests


Server Software:
Server Hostname:        localhost
Server Port:            3006

Document Path:          /features/feature1?scope=user-1
Document Length:        46 bytes

Concurrency Level:      20
Time taken for tests:   0.212 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      163000 bytes
HTML transferred:       46000 bytes
Requests per second:    4708.36 [#/sec] (mean)
Time per request:       4.248 [ms] (mean)
Time per request:       0.212 [ms] (mean, across all concurrent requests)
Transfer rate:          749.48 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   0.4      2       3
Processing:     0    2   1.3      2       9
Waiting:        0    2   1.3      2       9
Total:          1    4   1.4      4      11

Percentage of the requests served within a certain time (ms)
  50%      4
  66%      4
  75%      4
  80%      5
  90%      6
  95%      7
  98%      8
  99%      9
 100%     11 (longest request)
```
