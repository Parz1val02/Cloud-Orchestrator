#!/bin/bash

# parámetros
NombreVM="$1"
NombreOvS="$2"
VLAN_ID="$3"
PuertoVNC="$4"

generate_random_mac() {
	printf '20:20:20:20:%02X:%02X\n' $((RANDOM % 256)) $((RANDOM % 256))
}

puertoVNC_local="$((PuertoVNC - 5900))"

# crear interfaz TAP
interfaz_tap_vm="$NombreVM"-tap
ip tuntap add mode tap name "$interfaz_tap_vm"
random_mac=$(generate_random_mac)

# crear VM (script lab2)
qemu-system-x86_64 -enable-kvm -vnc 0.0.0.0:"$puertoVNC_local" -netdev tap,id="$interfaz_tap_vm",ifname="$interfaz_tap_vm",script=no,downscript=no -device e1000,netdev="$interfaz_tap_vm",mac="${random_mac}" -daemonize -snapshot cirros-0.6.2-x86_64-disk.img

# Conectar interfaz TAP al OvS del host local con el VLAN ID correspondiente
ovs-vsctl add-port "$NombreOvS" "$interfaz_tap_vm" tag="$VLAN_ID"
ip link set dev "$interfaz_tap_vm" up

# Mostrar información
echo "VM $NombreVM creada y conectada al OvS $NombreOvS con VLAN ID $VLAN_ID."
