# go-wharfie
Docker API client to enable the use of containers within Scientific Computing


## Example Run

```bash
$ docker-compose up
Starting ckniep_wharfie_1 ...
Starting ckniep_wharfie_1 ... done
Attaching to ckniep_wharfie_1
wharfie_1  | [II] qnib/init-plain script v0.4.28
wharfie_1  | > execute entrypoint '/opt/entry/00-logging.sh'
wharfie_1  | > execute entrypoint '/opt/entry/10-docker-secrets.env'
wharfie_1  | [II] No /run/secrets directory, skip step
wharfie_1  | > execute entrypoint '/opt/entry/99-remove-healthcheck-force.sh'
wharfie_1  | !!> Could not find specified ENTRYPOINTS_DIR '/opt/qnib/entry/'
wharfie_1  | > execute CMD '/opt/go-wharfie/bin/start.sh'
wharfie_1  | 2017/12/17 19:48:56 [II] Start Version: 0.0.0
wharfie_1  | Connected to 'ucp-controller-192.168.12.189' / v'ucp/2.2.4' (SWARM: active)
ckniep_wharfie_1 exited with code 0
```

## TODO

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

    - [ ] Fetch ssh command from mpirun
    - [ ] transform the ssh arguments to docker exec

