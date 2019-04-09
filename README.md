# kubic-control
Tool to deploy a kubernetes cluster on openSUSE Kubic using kubeadm and salt

## Design
Our plans are:

kubic-control consists of two binaries:
- kubicd, a daemon which communicates via gRPC with clients. It's setting up kubernetes on openSUSE Kubic, including pod network, kured, transactional-update, ...
- kubicctl, a cli interface

The communication is encrypted, the kubicctl command can run on any machine. The user authenticates with his certificate, a table specifies, which functions he is allowed to use. kubiccd will use kubeadm and kubectl to deploy and manage the cluster. So the admin can at everytime modify the cluster with this commands, too, there is no hidden state-database.  
