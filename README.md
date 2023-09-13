# kubectl-nsenter

Hey, buddy! Tired of the endless debug pods/node shells? `kubectl-nsenter` summoned to help you!

## Installation

```bash
kubectl krew install nsenter
```

## TL;DR

![nsenter demo](/img/demo.gif)

```bash
GLOBAL OPTIONS:
   --kubeconfig value                                       kubernetes client config path (default: $HOME/.kube/config) [$KUBECONFIG]
   --container value, -c value                              use namespace of specified container. By default first running container will taken
   --context value                                          override current context from kubeconfig
   --namespace value, -n value                              override namespace of current context from kubeconfig
   --user value, -u value                                   set username for ssh connection to node
   --password, -s                                           force ask for node password prompt (default: false)
   --ssh-auth-sock value                                    sets ssh-agent socket (default: current shell auth sock) [$SSH_AUTH_SOCK]
   --host value                                             override node ip
   --port value, -p value                                   sets ssh port
   --ns value [ --ns value ]                                define container's pid linux namespaces to enter. Sends transparently to nsenter cmd (default: "n")
   --interactive, -i                                        keep ssh session stdin (default: false)
   --tty, -t                                                allocate pseudo-TTY for ssh session (default: false)
   --ssh-opt value, -o value [ --ssh-opt value, -o value ]  same as -o for ssh client
   --use-node-name, -j                                      use kubernetes node name to connect with ssh. Useful with ssh configs (default: true) [$KUBECTL_NSENTER_USE_NODE_NAME]
   --help, -h                                               show help
   --version, -v                                            print the version
```

## What the kind is kubectl-nsenter?

`kubectl-nsenter` let you to exec to any pod's container linux namespace, such as network, mount etc. It uses a direct connection to node via ssh and supports two form of authentication: password and key. For auth by key it uses ssh-agent.

## How can I use this?

First we gotta talk about requirements:

- You **must** have a **root access** to node (with password or not) where pod is running
- Your client station **must** have ssh client binary in $PATH
- Your node **must** have CRI client for discovering container's pid (e.g. `crictl` for **containerd** or `docker` for **docker engine**)

If you can handle this requirements, we're moving on':

**Discover pod's opened tcp-ports**:

```bash
$ kubectl-nsenter -u vagrant httpbin-5876b4fbc9-rtvrq ss -tln
State         Recv-Q        Send-Q               Local Address:Port               Peer Address:Port       Process
LISTEN        0             128                        0.0.0.0:80                      0.0.0.0:*
```

**Discover pod's mounts**:

```bash
$ kubectl-nsenter -u vagrant --ns m --ns p  httpbin-5876b4fbc9-rtvrq mount -t xfs
/dev/vda1 on /dev/termination-log type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/resolv.conf type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/hostname type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/hosts type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
```

**Or start a full shell session as well**:

```bash
$ kubectl-nsenter -it httpbin-5876b4fbc9-rtvrq bash
[root@w-01 ~]#
```

Note, that ssh session requires keeping stdin (-i) and allocating pseudo-TTY (-t). Same as `docker run -it alpine sh`.

**Ultimate feature! Dump traffic from pod right on your station's wireshark!**

```bash
kubectl-nsenter postgres tcpdump -nnni any -w- | wireshark -ki-
```

## Init Containers

If desired pod is still initializing, nsenter will pick currently running container or fail, if none of init containers is running.

## Supported technologies

Container Runtimes Clients:

- docker;
- crictl - expected to be present on nodes with cri-o runtime;
- nerdctl - expected to be present on nodes with containerd runtime; crictl will be used as a fallback.

OS:

- Unix-like.
