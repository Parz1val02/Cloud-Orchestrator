#!/bin/bash

# parámetros
VLAN_ID="$1"
NombreOvS="$2"
NODE_ID="$3"
PROCESS_ID="$4"

ovs-vsctl del-port "$NombreOvS" "${NODE_ID}"vlan"${VLAN_ID}"-tap
ip link del "${NODE_ID}"vlan"${VLAN_ID}"-tap
kill -15 "${PROCESS_ID}"
