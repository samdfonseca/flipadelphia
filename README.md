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

```
$ mkdir ~/.flipadelphia
$ cp config/config.example.json ~/.flipadelphia/config.json
$ glide install
```

## Building

```
$ make build
```

## Running
```
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

```
$ curl -s localhost:3006/features/feature1?scope=user-1 | jq .
{
  "name": "feature1",
  "value": "on",
  "data": "true"
}
```

### Checking a feature

If the feature has been set on the scope, ```data``` is ```true```.
```
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
