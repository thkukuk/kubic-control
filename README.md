# kubic-control
[![license](http://img.shields.io/badge/license-apache_2.0-blue.svg?style=flat)](https://raw.githubusercontent.com/thkukuk/kubic-control/master/LICENSE)

Tool to deploy a kubernetes cluster on openSUSE Kubic using kubeadm and salt

kubic-control consists of two binaries:
- kubicd, a daemon which communicates via gRPC with clients. It's setting up kubernetes on openSUSE Kubic, including pod network, kured, transactional-update, ...
- kubicctl, a cli interface

The communication is encrypted, the kubicctl command can run on any machine. The user authenticates with his certificate, using RBAC to determine if the user is allowed to call this function. kubiccd will use kubeadm and kubectl to deploy and manage the cluster. So the admin can at everytime modify the cluster with this commands, too, there is no hidden state-database.

## Installation

`Kubicd`/`kubicctl` are using [salt](https://www.saltstack.com/) and [certstrap](https://github.com/square/certstrap) to manage nodes and certificates, additional kubeadm, kubectl, kubelet and crio have to be installed.
`Kubicd` has to run on the future kubernetes master node, `kubicctl` can run anywhere on the network. This requires only that `kubicd` is configured to listen on all interfaces, not only `localhost`. The salt minions have to be already accepted on the salt master. Before `kubicd` can be started, certificates have to be generated:

```
  kubicctl certificates initialize
```

This will create a CA and several certificates in `/etc/kubicd/pki`:
- Kubic-Control-CA.key - the private CA key
- Kubic-Control-CA.crt - the public CA key. This one is needed by kubicctl, too
- KubicD.key - the private key for kubicd
- kubicD.crt - the signed public key for kubicd
- admin.key - private key, allows kubicctl to connect to kubicd as admin
- admin.crt - public key, allows kubicctl to connect to kubcd as admin

For `kubicctl`, you need to create a directory `~/.config/kubicctl` which
contains `Kubic-Control-CA.crt`, `user.key` and `user.crt`. For the admin
role, this need to be a copy of admin.key and admin.crt. For other users,
you need to create corresponding certificates and sign them with
`Kubic-Control-CA.crt`. If you call `kubicctl` as root and there is no
`user.crt` in `~/.config/kubicctl`, the admin certificates from
`/etc/kubicd/pki` are used if they exist.
Certificates for additional users can be created with `kubicctl certificates
create <account>`.

Please take care of this certificates and store them secure, this are the
passwords to access kubicd!

## Deploy Kubernetes

To deploy the control-plane on the master with flannel as POD network and
`kured` to manage the reboot of nodes:

```
kubicctl init
```

For cilium instead of flannel you have to use `kubicctl init --pod-network cilium`.

To add additional nodes:

```
kubicctl node add node1,...
```

In the same way, you can remove nodes: `kubicctl node remove` or reboot
nodes: `kubicctl node reboot`.

To access with cluster with `kubectl`, you can get the kubeconfig with:
`kubicctl kubeconfig`.

The kubernetes cluster can be upgraded with:

```
kubicctl upgrade
```

## Configuration Files

`Kubicd` reads two configuration files: `kubicd.conf` and `rbac.conf`. The
first one is optional and contains the paths to the certificates and the server
name with port `kubicd` should listen to. The default file can be found in
`/usr/share/defaults/kubicd/kubicd.conf`. The variables can be overriden with
`/etc/kubicd/conf`, which only contains the changed entries.

The second file, `rbac.conf`, is mandatory, else nobody can access `kubicd`,
all requests will be rejected. The default file can be found in
`/usr/share/defaults/kubicd/rbac.conf`. Changed entries should be written
to `/etc/kubicd/rbac.conf`.

## RBAC

`rbac.conf` contains the roles as key and the users, who are allowed to use
this functionality as comma seperated list. `kubicctl rbac list` will print
out a list of current configured roles and the corresponding users. `kubicctl
rbac add <role> <user>` will add the user to the role.

## Usage

* certificates - Manage certificates for kubicd/kubicctl communication
  * create <user> - Create certificate for an user. The certificate will be stored in the local directory where you did call kubicctl.
  * initialize - Create CA, KubicD and admin certificates. This certificates will be stored in `/etc/kubicd/pki/`
* help - Help about any command
* init - Initialize Kubernetes Master Node
  * --pod-network=<flannel|cilium>	Pod network
  * --adv-addr=<IPaddr>	IP address the API Server will advertise on
*  kubeconfig - Download kubeconfig
  * --output=<file> - Where the kubeconfig file should be stored	
* node - Manage kubernetes nodes
  * add <node>,... - Add new nodes to cluster. Node names must be the name used by salt for that node. A comma seperated list or '[]' syntax are allowed to specify more than one new node.
  * list - List all reacheable worker nodes
  * reboot <node> - Reboot node from cluster. Node will be drained first. Node name must be the name used by salt for that node.
  * remove - Remove node form cluster
* rbac - Manage RBAC rules
  * add <role> <user> - Add user account to a role
  * list - List roles and accounts
* upgrade - Upgrade Kubernetes Cluster to the version of the installed kubelet command
* version - Print version information


## Notes

`Kubicd` does not store any informations about the state of the kubernetes
cluster. This allows to manage the cluster with `kubectl` and `kubeadm`
yourself without `kubicctl`. There is only one important thing: a grain
has to be set on new worker nodes: `kubicd=kubic-worker-node`. If nodes
gets manual removed, this grain has to be deleted, too.
