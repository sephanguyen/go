# RKE2-Ansible-HA
Example
----------------

- Setup rke2 cluster with high availability
```
ansible-playbook site.yml -i inventory/rke2/hosts.ini --diff
```
- To get access to Kubernetes cluster
```
/var/lib/rancher/rke2/bin/kubectl --kubeconfig /etc/rancher/rke2/rke2.yaml get nodes
```
- To uninstall a node
```
/usr/local/bin/rke2-uninstall.sh
```
