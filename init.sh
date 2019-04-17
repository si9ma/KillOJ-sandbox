#!/usr/bin/env bash

mount -oremount,rw /sys/fs/cgroup
kbox $@
