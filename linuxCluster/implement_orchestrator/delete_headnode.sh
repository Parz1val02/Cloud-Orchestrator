#!/bin/bash

# parámetros
VLAN_ID="$1"
NombreOvS="$2"

# borrar puertos en ovs
ovs-vsctl del-port "$NombreOvS" vlan"${VLAN_ID}"
ovs-vsctl del-port "$NombreOvS" vlan"${VLAN_ID}"-veth1

ip link del vlan"${VLAN_ID}"-veth1
ip netns del vlan"${VLAN_ID}"-dhcp
