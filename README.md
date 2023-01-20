# sminit

- sminit is a trivial service manager on linux trying to mimic [zinit](https://github.com/threefoldtech/zinit).
- it provides the functionality to start, stop, add, delete, and show services' logs.
- it allows service dependency, e.g. you can tell it to start service a when service b is in running status.
- users are able to specify whether they want to run a service once, or indefinitely.

## Usage

- download the binary:
  
  ```bash
  wget https://github.com/MarioBassem/sminit/releases/download/v0.0.2/sminit
  ```

- create service definition files in `/etc/sminit`
- run ```sminit swatch``` with root user privileges to tell sminit to keep track of services in `/etc/sminit` and start whichever is eligible.
- to add a new service to tracked services, create its definition file in `/etc/sminit/example_service.yaml`, then run ```sminit add example_service```.
- to delete a service from tracked services, run ```sminit delete example_service```.
- to start a stopped service, run ```sminit start example_service```.
- to stop a started or running service, run ```sminit stop example_service```.
- to show sminit logs, run ```sminit log```.

## Creating a service definition file

- Services should be defined in yaml files in `/etc/sminit`.
- A service definition file has the following fields:
  
  - `cmd`: this is the command that is executed when the service is eligible to run.
  - `log`: if this is equal to "stdout", sminit will dump the logs of this service with sminit's logs, available with `sminit log`
  - `after`: this is a list of the services that should be in a running state before sminit starts this service.
  - `oneshot`: this is a boolean flag indicating whether to keep starting this service if it is terminated, or run it only once.
  - `healthcheck`: this is a command that has to successfuly run before declaring this service as running. the default is `sleep 1`.
