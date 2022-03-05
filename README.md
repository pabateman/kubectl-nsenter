# kubectl-nsenter

Hey, buddy! Tired of the endless debug pods/node shells? `kubectl-nsenter` summoned to help you!
## TL;DR

![nsenter demo](/img/demo.gif)


```
GLOBAL OPTIONS:
   --kubeconfig value           kubernetes client config path (default: $HOME/.kube/config) [$KUBECONFIG]
   --container value, -c value  use namespace of specified container. By default first container will taken
   --context value              override current context from kubeconfig
   --namespace value, -n value  override namespace of current context from kubeconfig
   --user value, -u value       set username for ssh connection to node (default: "johndoe") [$USER]
   --password, -s               force ask for node password prompt (default: false)
   --ssh-auth-sock value        sets ssh-agent socket (default: current shell auth sock) [$SSH_AUTH_SOCK]
   --port value, -p value       sets ssh port (default: "22")
   --ns value                   define container's pid linux namespaces to enter. sends transparently to nsenter cmd (default: "n")
   --help, -h                   show help (default: false)
   --version, -v                print the version (default: false)

```

## What the kind is kubectl-nsenter?

`kubectl-nsenter` let you to exec to any pod's container linux namespace, such as network, mount etc. It uses a direct connection to node via ssh and supports two form of authentication: password and key. For auth by key it uses ssh-agent.

## How can i use this?

Discover pod's opened tcp-ports:

```
$ kubectl-nsenter -u vagrant httpbin-5876b4fbc9-rtvrq ss -tln
State         Recv-Q        Send-Q               Local Address:Port               Peer Address:Port       Process
LISTEN        0             128                        0.0.0.0:80                      0.0.0.0:*
```

Discover pod's mounts:

```
$ kubectl-nsenter -u vagrant --ns m --ns p  httpbin-5876b4fbc9-rtvrq mount -t xfs
/dev/vda1 on /dev/termination-log type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/resolv.conf type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/hostname type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
/dev/vda1 on /etc/hosts type xfs (rw,relatime,seclabel,attr2,inode64,logbufs=8,logbsize=32k,noquota)
```

And so on!

## Supported technologies

SSH:

- Ssh-agent;
- Password.

Container Runtimes:

- Docker;
- Containerd.

OS:

- Unix-like.

## Known limitations

- Unfortunately, there are only interactive session with tty allocating available. Will be fixed in future versions.
- Currently there is no way to use this plugin on Windows because tty issues.
