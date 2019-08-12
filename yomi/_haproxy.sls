# Meta pillar for Yomi
#
# There are some parameters that can be configured and adapted to
# launch a basic Yomi installation:
#
#   * efi = {True, False}
#   * baremetal = {True, False}
#   * disk = {/dev/...}
#   * repo-main = {https://download....}
#
# This meta-pillar can be used as a template for new installers. This
# template is expected to be adapted for production systems, as was
# designed for CI / QA and development.

config:
  events: no
  reboot: no
  snapper: yes
  grub2_theme: yes
{% if efi %}
  grub2_console: yes
{% endif %}
  locale: en_US.UTF-8
  keymap: us
  timezone: UTC
  hostname: node

#
# Storage section for a microos deployment in a single device
#

partitions:
  config:
    label: gpt
  devices:
    {{ disk }}:
      initial_gap: 1MB
      partitions:
{% if not efi %}
        - number: 1
          size: 1MB
          type: boot
{% else %}
        - number: 1
          size: 256MB
          type: efi
{% endif %}
        - number: 2
          size: 16384MB
          type: linux
        - number: 3
          size: rest
          type: linux

filesystems:
{% if efi %}
  {{ disk }}1:
    filesystem: vfat
    mountpoint: /boot/efi
{% endif %}
  {{ disk }}2:
    filesystem: btrfs
    mountpoint: /
    options: [ro]
    subvolumes:
      prefix: '@'
      subvolume:
        - path: root
        - path: tmp
        - path: home
        - path: opt
        - path: srv
        - path: boot/writable
        - path: usr/local
        - path: boot/grub2/i386-pc
        - path: boot/grub2/x86_64-efi
  {{ disk }}3:
    filesystem: btrfs
    mountpoint: /var

bootloader:
  device: {{ disk }}
  kernel: swapaccount=1
  disable_os_prober: yes

software:
  config:
    minimal: yes
  repositories:
    repo-main: {{ repo_main }}
  packages:
    - pattern:microos_base
    - pattern:microos_defaults
{% if baremetal %}
    - pattern:microos_hardware
{% else %}
    - kernel-default-base
{% endif %}
    - haproxy

salt-minion:
  configure: yes

services:
  enabled:
    - salt-minion

users:
  - username: root
    # Set the password as 'linux'. Do not do that in production
    password: "$1$wYJUgpM5$RXMMeASDc035eX.NbYWFl0"
    # Personal certificate, without the type prefix nor the host
    # suffix
    certificates:
      - "AAAAB3NzaC1yc2EAAAADAQABAAABAQDdP6oez825gnOLVZu70KqJXpqL4fGf\
        aFNk87GSk3xLRjixGtr013+hcN03ZRKU0/2S7J0T/dICc2dhG9xAqa/A31Qac\
        hQeg2RhPxM2SL+wgzx0geDmf6XDhhe8reos5jgzw6Pq59gyWfurlZaMEZAoOY\
        kfNb5OG4vQQN8Z7hldx+DBANPbylApurVz6h5vvRrkPfuRVN5ZxOkI+LeWhpo\
        vX5XK3eTjetAwWEro6AAXpGoQQQDjSOoYHCUmXzcZkmIWEubCZvAI4RZ+XCZs\
        +wTeO2RIRsunqP8J+XW4cZ28RZBc9K4I1BV8C6wBxN328LRQcilzw+Me+Lfre\
        eDPglqx"
