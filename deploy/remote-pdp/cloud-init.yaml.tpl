#cloud-config

package_update: true
packages:
  - mesa-opencl-icd
  - ocl-icd-opencl-dev
  - gcc
  - git
  - jq
  - pkg-config
  - curl
  - clang
  - build-essential
  - hwloc
  - libhwloc-dev
  - wget
  - aria2
  - pgcli
  - python-is-python3
  - ccze

write_files:
%{ for wf in write_files ~}
  - path: ${wf.path}
    permissions: "${wf.permissions}"
    encoding: base64
    content: ${base64encode(wf.content)}
%{ endfor }

runcmd:
%{ for cmd in runcmd_steps ~}
  - ${cmd}
%{ endfor }
