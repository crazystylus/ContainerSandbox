# Sandbox
## Description
This demonstarates deployment of a stable sandbox inside a running container for running untrusted codes and applications <br>
If is meant to be used as a code judge base sandbox to be run inside a kubernetes pod or in a docker container

## Breakdown
1. Namespaces :- It uses the following namespace -> UTS Namespace, Mount Namespace, IPC Namespace, PID Namespace and a Network Namespace
2. CGroups :- It uses cpu, memory and pids cgroups to cut down fork bombs and memory and cpu eating malicious codes
3. UnPriviledged user :- An unpriviledged user is used for compilation and execution of the programs
4. *EXTRA* chroot :- Chroot support is there in case required, but it then needs a rootfs to switch to

## Usage
Primarily it was tested on Podman v1.5.1 <br>
Copy Files to the git repo to a folder or pull in the container <br>
<br>
<pre>
> podman run -it --name gochk --cap-add=SYS_ADMIN -v /sandbox:/sandbox golang:alpine
> apk add openrc gcc libc-dev bash
> mkdir proc
> adduser sandbox # (uid and gid should be 1000 for this user)
> go build -o sandbox
> ./sandbox run 60 /bin/sh
</pre>
## Benchmark
### Adds 8ms latency per execution