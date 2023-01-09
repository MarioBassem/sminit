# sminit

- sminit is a trivial service manager.
- it provides the functionality to start, stop, add, delete, and show the logs of services.
- it allows service dependency, e.g. you can tell it to start service a when service b is in running status.
- users are able to specify whether they want to run a service once, or indefinitely.

## Commands

- swatcher:
    long living process. it is run only once when the system boots. it is responsible for:
  - keeping running services alive.
  - starting pending services.
  - managing services' logs.
  - managing services' orders.
  
- add [service]:
    responsible for:
  - loading service config file.
  - providing `swatcher` with the proper information needed to manage the service.

- start [service]:
    responsible for:
  - sending a request for `swatcher` to start this service.

- stop [service]:
    responsible for:
  - sending a request for `swatcher` to stop this service.
  
- delete [service]:
    responsible for:
  - sending a request for `swatcher` to delete this service.

- log [service]:
    responsible for:
  - showing the logs for all services
