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
`Kubic-Control-CA.crt`.
Please take care of this certificates and store them secure, this are the
passwords to access kubicd!

## Usage

To deploy the control-plane on the master with flannel as POD network and
`kured` to manage the reboot of nodes:

```
kubicctl init
```

To add additional nodes:

```
kubicctl node add
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
`/usr/share/defaults/kubicd/kubicd.conf`. Changed entries should be written
to `/etc/kubicd/rbac.conf`.

## Notes

`Kubicd` does not store any informations about the state of the kubernetes
cluster. This allows to manage the cluster with `kubectl` and `kubeadm`
yourself without `kubicctl`. There is only one important thing: a grain
has to be set on new worker nodes: `kubicd=kubic-worker-node`. If nodes
gets manual removed, this grain has to be deleted, too.
