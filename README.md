# sealctl for what?

sealctl is kubernetes multi tencent command line tool.

Using cases:

1. Generate kubeconfig file for a nomal user, like a developer that we don't want him has privilege admin access.
2. Group manage, different group have different permissions can access different kubernetes namespaces.
3. Manage roles...
4. Namespace Quota..

# Quick start

Create a user named fanux, and join in two group sealyun and sealos

```
sealctl user -u fanux --group sealyun --group sealos
```
Then sealctl will generate a kubeconfig for fanux.

```
$ cat kubeconfig
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBD...
    server: https://sealyun.com:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: fanux
  name: fanux@kubernetes
current-context: fanux@kubernetes
kind: Config
preferences: {}
users:
- name: fanux
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDR...
    client-key-data: LS0tLS1CRUd...
```
fanux has no access to pods before we bind a role to he.

```
# kubectl --kubeconfig kubeconfig get pod
Error from server (Forbidden): pods is forbidden: User "fanux" cannot list resource "pods" in API group "
```

> Bind a role for user or group

You can bind role to user or group.

Set fanux as cluster admin..

```
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: user-admin-test
subjects:
- kind: User
  name: "fanux" # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin  # using admin role
  apiGroup: rbac.authorization.k8s.io
```
All users in group sealos has admin authority

```
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: group-admin-test
subjects:
- kind: Group
  name: "sealos" # Name is case sensitive
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin  # using admin role
  apiGroup: rbac.authorization.k8s.io
```

# Command Reference

```shell script
./sealctl user -h
Easy to use this to create a kubernetes user, 
           if your want some one access your kubernetes cluster read only, 
           you can use this command generate a kubeconfig for him, and bind 
           read only role etc..

Usage:
  sealctl user [flags]

Flags:
  -s, --apiserver string      apiserver address (default "https://apiserver.cluster.local:6443")
      --ca-crt string         kubernetes ca crt file (default "/etc/kubernetes/ca.crt")
      --ca-key string         kubernetes ca key file (default "/etc/kubernetes/ca.key")
      --cluster-name string   kubeconfig cluster name (default "kubernetes")
  -d, --dns strings           apiserver certSANs dns list (default [apiserver.cluster.local,localhost,sealyun.com])
  -g, --group strings         user group names (default [sealyun,alibaba])
  -h, --help                  help for user
      --ips strings           apiserver certSANs ip list (default [127.0.0.1,10.103.97.2])
  -o, --out string            default kube config out put file name (default "./kube/config")
  -u, --user string           user name in your kube config (default "fanux")
```
