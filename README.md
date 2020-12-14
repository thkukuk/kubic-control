# kubic-control
[![license](http://img.shields.io/badge/license-apache_2.0-blue.svg?style=flat)](https://raw.githubusercontent.com/thkukuk/kubic-control/master/LICENSE)

Tool to deploy a kubernetes cluster on openSUSE Kubic using kubeadm and salt

kubic-control consists of two binaries:
- kubicd, a daemon which communicates via gRPC with clients. It's setting up kubernetes on openSUSE Kubic, including pod network, kured, transactional-update, ...
- kubicctl, a cli interface

The communication is encrypted, the kubicctl command can run on any machine. The user authenticates with his certificate, using RBAC to determine if the user is allowed to call this function. kubiccd will use kubeadm and kubectl to deploy and manage the cluster. So the admin can modify the cluster with this command, too. There is no hidden state-database except for the informations necessary for a kubernetes multi-master/HA setup.

## Requirements

Mainly generic requirements by kubernetes itself:

- All the nodes on the cluster must be on a the same network and be able to communicate directly with each other.
- All nodes in the cluster must be assigned static IP addresses. Using dynamically assigned IPs will break cluster functionality if the IP address changes.
- The Kubernetes master node(s) must have valid Fully-Qualified Domain Names (FQDNs), which can be resolved both by all other nodes and from other networks which need to access the cluster.
- Since Kubernetes mainly works with certificates and tokens, the time on all nodes needs to be in sync. Otherwise communication inside the cluster will break.


## Installation

`Kubicd`/`kubicctl` are using [salt](https://www.saltstack.com/) and [certstrap](https://github.com/square/certstrap) to manage nodes and certificates. Additionally kubeadm, kubectl, kubelet and crio have to be installed.
`Kubicd` has to run on the salt master host. If there is not already a salt-master in the network, kubicd and the salt-master can run on the future kubernetes master node. `kubicctl` can run anywhere on the network. This requires only that `kubicd` is configured to listen on all interfaces, not only `localhost`. The salt minions have to be already accepted on the salt master.

Before `kubicd` can be started, certificates have to be generated. Starting and enabling the service `kubicd-init` takes care of that.

This will create a CA and several certificates in `/etc/kubicd/pki`:
- Kubic-Control-CA.key - the private CA key
- Kubic-Control-CA.crt - the public CA key. This one is needed by kubicctl, too
- KubicD.key - the private key for kubicd
- kubicD.crt - the signed public key for kubicd
- admin.key - private key, allows kubicctl to connect to kubicd as admin
- admin.crt - public key, allows kubicctl to connect to kubcd as admin

For `kubicctl`, you need to create a directory `~/.config/kubicctl` which
contains `Kubic-Control-CA.crt`, `user.key` and `user.crt`. For the admin
role, this need to be a copy of `admin.key` and `admin.crt`. For other users,
you need to create corresponding certificates and sign them with
`Kubic-Control-CA.crt`.
If you call `kubicctl` as root and there is no
`user.crt` in `/root/.config/kubicctl`, the admin certificates from
`/etc/kubicd/pki` are used if they exist.
Certificates for additional users can be created with `kubicctl certificates
create <account>`.

Please take care of these certificates and store them secure, these are the
passwords to access kubicd!

## Deploy Kubernetes

The first question is: single master node or high-availability kubernetes
masters? In the first case, there is not much to do: you need one master
machine, which is also running `kubicd`, and that's it.
If you want a high-availability kubernetes master, you need three machines
which meet kubeadm's minimum requirements for masters. Additional, you need a
load balancer. The load balancer must be able to communicate with all control
plane nodes on the apiserver port. It must also allow incoming traffic on its
listening port 6443. If you have no loadbalancer, you can use HAProxy. This load balancer
is only for the kubernetes control-plane. For deployments, something like
`metallb` is still needed.

If you installed the Kubic Admin Node system role, kubicd-init and kubicd should be enabled by default. If you installed using an image, you need to manually start and enable them:
```bash
systemctl enable --now kubicd-init
systemctl enable --now kubicd
```

To deploy the control-plane on the master with weave as POD network and
`kured` to manage the reboot of nodes:

```
kubicctl init
```

To deploy a highly available master, the DNS name of the load balancer needs
to be specified. The cluster will be reacheable under this DNS name.

```
kubicctl init --multi-master load.balancer.dns
```

If the haproxy is also a salt-minion and should be configured
and adjusted automatically:

```
kubicctl init --haproxy salt-minion --multi-master load.balancer.dns
```

In this case, kubicd will configure the haproxy and add or remove master nodes
depeding on the kubernetes cluster configuration automatically, if `haproxycfg`
is installed.

For cilium or flannel instead of weave you have to use `kubicctl init
--pod-network cilium` or `kubicctl init --pod-network flannel`.

To deploy kubic without a CNI you have to use `kubicctl init 
--pod-network none`

To add additional worker nodes:

```
kubicctl node add node1,...
```

In the high-availability case, two additional masters need to be added:

```
kubicctl node add --type master master2
```
```
kubicctl node add --type master master3
```

Make sure the loadbalancer can reach all three master nodes.


In the same way as new nodes were added, existing nodes can also be removed:
`kubicctl node remove` or rebooted: `kubicctl node reboot`. Please make
sure that you always have three master nodes in case of high-availbility masters.

To access the cluster with `kubectl`, you can get the kubeconfig with:
`kubicctl kubeconfig`.

The kubernetes cluster can be upgraded with:

```
kubicctl upgrade
```

## Configuration Files

`kubicd` reads two configuration files: `kubicd.conf` and `rbac.conf`. The
first one is optional and contains the paths to the certificates and the server
name and port that `kubicd` should listen to. The default file can be found in
`/usr/etc/kubicd/kubicd.conf`. The variables can be overriden with
`/etc/kubicd/kubicd.conf`, which only needs to contain the changed entries.

The second file, `rbac.conf`, is mandatory, else nobody can access `kubicd` and
all requests will be rejected. The default file can be found in
`/usr/etc/kubicd/rbac.conf`. Changed entries should be written
to `/etc/kubicd/rbac.conf`.

`kubicctl` optionally reads a `~/config/kubicctl/kubicctl.conf`, which
allows you to configure the hostname and port of a remote `kubicd` process:

```
  [global]
  server = remote.host.name
  port = 7148
```

## RBAC

`rbac.conf` contains the roles as key and the users, who are allowed to use
this functionality, as a comma separated list. `kubicctl rbac list` will print
out a list of currently configured roles and the corresponding users. `kubicctl
rbac add <role> <user>` will add the user to the role.

## Deploy new nodes

`kubicd` has support to deploy new nodes with help of
[Yomi](https://github.com/openSUSE/yomi). This requires salt pillars for the
new node describing what should be installed where . There is a `kubicctl node
deploy prepare` command, which will create the required pillars based on the
type of the new node and the hardware. After this step, the generated pillars
should be verified, if really the right things will be deleted and installed.
Afterwards, `kubicctl node deploy install` will erase the content of the
harddisk, install the new node and, if this new node is of type "master" or
"worker", will also add it to the kubernetes cluster.

## Usage

* certificates - Manage certificates for kubicd/kubicctl communication
  * create <user> - Create certificate for an user. The certificate will be stored in the local directory where you did call kubicctl.
  * initialize - Create CA, KubicD and admin certificates. This certificates will be stored in `/etc/kubicd/pki/`
* help - Help about any command
* init - Initialize Kubernetes Master Node
  * `--multi-master=<DNS name>`  	Setup HA masters, the argument must be the DNS name of the load balancer
  * `--haproxy=<salt name>` Adjust haproxy configuration for multi-master setup via salt
  * `--pod-network=<flannel|cilium>`	Pod network
  * `--adv-addr=<IPaddr>`	IP address the API Server will advertise on
  * `--apiserver_cert_extra_sans=<IPaddr>`	additional IPs to add to the APIserver certificate
  * `--stage=<official|devel>` Specify to use the official images or from the devel project
* kubeconfig - Download kubeconfig
  * `--output=<file>` - Where the kubeconfig file should be stored
* node - Manage kubernetes nodes
  * add <node>,... - Add new nodes to cluster. Node names must be the name used by salt for that node. A comma separated list or '[]' syntax are allowed to specify more than one new node.
  * list - List all reacheable worker nodes
  * reboot <node> - Reboot node from cluster. Node will be drained first. Node name must be the name used by salt for that node.
  * remove - Remove node form cluster
  * deploy - Install a new node
    * prepare <type> <node> - Prepare configuration to install new node with Yomi
    * install <type> <node> - Install new node with Yomi
* deploy - Install a new service
  * hello-kubic - Install a hello kubic demo webservices
  * metallb - Install the MetalLB loadbalancer
* rbac - Manage RBAC rules
  * add <role> <user> - Add user account to a role
  * list - List roles and accounts
* upgrade - Upgrade Kubernetes Cluster to the version of the installed kubeadm command if not otherwise specified
* destroy-cluster - Remove all worker and master nodes
* status - Print status informations of KubicD
* version - Print version information

## Backup

On the machine where `kubicd` is running, `/etc/kubicd` and
`/var/lib/kubic-control` should be part of the backup.

## Notes

`Kubicd` does not store any informations about the state of the kubernetes
cluster except for the deployed daemonsets. This allows to manage the cluster
with `kubectl` and `kubeadm` yourself without `kubicctl`. Daemonsets not
installed via `kubicctl`/`kubicd` have to be updated by the admin themself,
they will not be updated by `kubicctl upgrade`.

There is only one important thing: a grain has to be set on new worker and
master nodes:
- `kubicd=kubic-worker-node` for worker nodes
- `kubicd=kubic-master-node` for additional master nodes

If a node
gets removed manually, this grain has to be deleted, too.
