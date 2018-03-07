# go-wharfie
Docker API client to enable the use of containers within Scientific Computing


## Roadmap

- [X] Create basic CLI

    - [X] Boilerplate CLI

- [X] DockerEE interaction

    - [X] Connect to DockerEngine using ClientBundle
    - [X] Add/rm `job.id=<jobid>` label to all nodes within the job
    - [X] Create Service
    - [X] Wait for service to be up
    - [X] Get ContainerID of all tasks
    - [X] Destroy service
    - [X] Wait for all container of service to be destroyed

- [ ] HPC interactions

    - [X] Fetch ssh command from mpirun
    - [X] transform the ssh arguments to docker exec bash command
    - [ ] use docker client to drop `docker exec` fork

## Todo

- [ ] Remove network via `go-wharfie remove`
