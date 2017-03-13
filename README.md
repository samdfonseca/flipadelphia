# Flipadelphia

*Flipadelphia flips your features*

[![Build Status](https://travis-ci.org/samdfonseca/flipadelphia.svg?branch=master)](https://travis-ci.org/samdfonseca/flipadelphia)

<img src="http://i.imgur.com/28TTvje.gif" alt="flipadelphia"/>

## Installation
* Flipadelphia uses a config file to keep track of settings for runtime environments. By default, the file is expected
    to be in the ~/.flipadelphia directory. You can also manually set the config with the -c/--config flag. An example
    of the config file is in the config subpackage.
* To handle dependencies, Flipadelphia uses the [dep](https://github.com/golang/dep) dependency manager. Dep
    installs dependencies into the ```vendor``` directory within a project, and handles versioning of dependencies.

```sh
$ mkdir ~/.flipadelphia
$ cp config/config.example.json ~/.flipadelphia/config.json
$ ./Taskfile deps
```

## Building

```sh
$ ./Taskfile build
```

## BoltDB Data Layout

3 top level buckets
- features
- scopes
- values

"features" bucket
- feature1 [bucket]
-- scope1: "uuid1"
-- scope2: "uuid2"

"scopes" bucket
- scope1 [bucket]
-- feature1: "uuid1"
- scope2 [bucket]
-- feature1: "uuid2"

"values" bucket
- uuid1: "on"
- uuid2: "off"

## Running
```sh
$ ./flipadelphia help
NAME:
   flipadelphia - Start the Flipadelphia server

USAGE:
   flipadelphia [global options] command [command options] [arguments...]

VERSION:
   dev-build

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --env value, -e value  An environment from the config.json file to use (default: "bolt") [$FLIPADELPHIA_ENV]
   --config value         Path to the config file. (default: "config.json") [$FLIPADELPHIA_CONFIG]
   --help, -h             show help
   --version, -v          print the version

$ ./flipadelphia
flipadelphia: Using BoltDB persistence store: /Users/samfonseca/.flipadelphia/flipadelphia_dev.db
flipadelphia: Listening on port 3006
```

## Usage

### Setting a feature

Auth settings are defined in the config.json file for the runtime environment

```sh
$ curl -s -d '{"scope":"user-1","value":"on"}' -X POST localhost:3006/admin/features/feature1 | jq .
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
$ curl -s localhost:3006/features/unset_feature?scope=user-1 | jq .
{
  "name": "unset_feature",
  "value": "",
  "data": "false"
}
```

### Checking a scope

Get all features set on a scope

```sh
$ curl -s localhost:3006/features?scope=user-1 | jq .
[
  "feature1"
]
```

Get features set on a scope matching some value

```sh
$ curl -s localhost:3006/features?scope=user-1\&value=off | jq .
[]
$ curl -s -d '{"scope":"user-1","value":"1"}' -X POST localhost:3006/features/feature2 | jq .
{
  "name": "feature2",
  "value": "1",
  "data": "true"
}
$ curl -s localhost:3006/features?scope=user-1\&value=1 | jq .
[
  "feature2"
]
```

### Admin functions

Get all scopes

```sh
$ curl -s localhost:3006/admin/scopes | jq .
[
  "user-1",
  "user-2",
  "user-3",
  "venue-1",
  "venue-2",
  "venue-3"
]
```

Get all scopes matching a prefix

```sh
$ curl -s localhost:3006/admin/scopes?prefix=user | jq .
[
  "user-1",
  "user-2",
  "user-3"
]
```

Get all features

```sh
$ curl -s localhost:3006/admin/features | jq .
[
  "feature1",
  "feature2",
  "feature3"
]
```

## Performance

* Flipadelphia uses BoltDB as the persistence layer. BoltDB fits the nature of a feature flipping service because it's a read-optimized database and features are typically checked far more often than set.
* To keep the response time as low as possible, all the check endpoints come without authorization. The thinking here is that theres no harm in someone checking a feature. If thats an issue, use a uuid for the feature and scope and keep a mapping of those externally. Authentication will be added to the set endpoint soon.

```sh
$ ab -n 10000 -c 20 localhost:3006/features/feature1?scope=user-1
This is ApacheBench, Version 2.3 <$Revision: 1663405 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:
Server Hostname:        localhost
Server Port:            3006

Document Path:          /features/feature1?scope=user-1
Document Length:        46 bytes

Concurrency Level:      20
Time taken for tests:   2.051 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1630000 bytes
HTML transferred:       460000 bytes
Requests per second:    4875.25 [#/sec] (mean)
Time per request:       4.102 [ms] (mean)
Time per request:       0.205 [ms] (mean, across all concurrent requests)
Transfer rate:          776.04 [Kbytes/sec] received

Connection Times (ms)
            min  mean[+/-sd] median   max
Connect:        0    2   5.7      2     154
Processing:     0    2   4.6      2     154
Waiting:        0    2   3.8      2     153
Total:          1    4   7.3      4     155

Percentage of the requests served within a certain time (ms)
50%      4
66%      4
75%      4
80%      4
90%      5
95%      5
98%      6
99%      8
100%    155 (longest request)


$ ab -n 10000 -c 20 localhost:3006/features?scope=user-1\&value=on
This is ApacheBench, Version 2.3 <$Revision: 1663405 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 1000 requests
Completed 2000 requests
Completed 3000 requests
Completed 4000 requests
Completed 5000 requests
Completed 6000 requests
Completed 7000 requests
Completed 8000 requests
Completed 9000 requests
Completed 10000 requests
Finished 10000 requests


Server Software:
Server Hostname:        localhost
Server Port:            3006

Document Path:          /features?scope=user-1&value=on
Document Length:        34 bytes

Concurrency Level:      20
Time taken for tests:   1.538 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      1510000 bytes
HTML transferred:       340000 bytes
Requests per second:    6501.95 [#/sec] (mean)
Time per request:       3.076 [ms] (mean)
Time per request:       0.154 [ms] (mean, across all concurrent requests)
Transfer rate:          958.78 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   3.3      1     112
Processing:     1    2   3.7      1     112
Waiting:        1    2   3.7      1     112
Total:          1    3   5.0      3     114

Percentage of the requests served within a certain time (ms)
  50%      3
  66%      3
  75%      3
  80%      3
  90%      3
  95%      4
  98%      4
  99%      4
 100%    114 (longest request)
```
