---
name: Bug Report
about: Report a bug encountered while using Network Operator
labels: bug

---

<!-- Please use this template while reporting a bug and provide as much info as possible.
-->


**What happened**:

**What you expected to happen**:

**How to reproduce it (as minimally and precisely as possible)**:

**Anything else we need to know?**:

**Logs**:
- NicClusterPolicy CR spec and state:
- Output of: `kubectl -n nvidia-network-operator get -A`:
- Network Operator version:
- Logs of Network Operator controller:
- Logs of the various Pods in `nvidia-network-operator` namespace:
- Helm Configuration (if applicable):
- Kubernetes' nodes information (labels, annotations and status): `kubectl get node -o yaml`:

**Environment**:
- Kubernetes version (use `kubectl version`): 
- Hardware configuration:
  - Network adapter model and firmware version:
- OS (e.g: `cat /etc/os-release`):
- Kernel (e.g. `uname -a`):
- Others:
