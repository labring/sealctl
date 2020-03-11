# sealctl

Kubernetes multi-tencent command line tools, manage user, authorization,quota etc..

## User and group

Create a user named fanux, and join in two group sealyun and sealos
```
sealctl user create fanux --group sealyun --group sealos
```

Then sealctl will generate a kubeconfig for `fanux`.

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

`fanux` has no access to pods before we bind a role to he.
```
# kubectl --kubeconfig kubeconfig get pod
Error from server (Forbidden): pods is forbidden: User "fanux" cannot list resource "pods" in API group "" in the namespace "default"
```

### Group

Only group admin can bind user to his group.

fanux create a group and set Bob as admin, Alice is member, who create the group his default role is admin, if group sealyun not exist,will create it.
```shell
sealctl group apply sealyun --admin Bob --member Alice
```
Then you can see Group CRD:
```shell script
kubectl get group
```

## Bind user or group to namespace

```
# forbidden to default
kubectl --kubeconfig kubeconfig get pod -n default 
sealctl bind --user fanux --ns default
# fanux access to default now
kubectl --kubeconfig kubeconfig get pod -n default

# forbidden to kube-system
kubectl --kubeconfig kubeconfig get pod -n kube-system
sealctl bind --group sealos --ns kube-system
# access to kube-system now
kubectl --kubeconfig kubeconfig get pod -n kube-system
```

## Bind a role

You can bind role to user or group.

Set `fanux` as cluster admin..
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

All users in group `sealos` has admin authority
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

## Set namespace quota

```
sealctl quota --ns default --CPU 4 --memory 5G --disk 200G
```

## namespace role

Assign read write and delete access to fanux for namespace kube-system
```shell
sealctl chmod -rwd --user fanux --namespace kube-system
```

See clusterrole and clusterrolebinding:
```shell script
kubectl get clusterrole
kubectl get clusterrolebinding
```
