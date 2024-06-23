#!/bin/bash

# parámetros
NombreRed="$1"
VLAN_ID="$2"
DireccionRed="$3"   # formato CIDR
RangoDHCP="$4"      # especificar formato
brigde="br-int"

echo $NombreRed
echo $VLAN_ID
echo $DireccionRed
echo $RangoDHCP


# Extraer la dirección de red y la máscara de subred
direccion_red=$(echo $DireccionRed | cut -d '/' -f1)
mascara_subred=$(echo $DireccionRed | cut -d '/' -f2)

# Calcular la primera dirección disponible
IFS='.' read -r -a octetos <<< "$direccion_red"
octetos[3]=$(( ${octetos[3]} + 1 ))
primera_direccion_disponible_cidr="${octetos[0]}.${octetos[1]}.${octetos[2]}.${octetos[3]}/$mascara_subred"
primera_direccion_disponible_sincdr="${octetos[0]}.${octetos[1]}.${octetos[2]}.${octetos[3]}"

# Calcular la segunda dirección disponible sumando 1 a la primera dirección disponible
IFS='.' read -r -a octetos_segunda <<< "$primera_direccion_disponible_sincdr"
octetos_segunda[3]=$(( ${octetos_segunda[3]} + 1 ))
segunda_direccion_disponible_cidr="${octetos_segunda[0]}.${octetos_segunda[1]}.${octetos_segunda[2]}.${octetos_segunda[3]}/$mascara_subred"
echo "La primera dirección disponible es: $primera_direccion_disponible_cidr"
echo "La segunda dirección disponible es: $segunda_direccion_disponible_cidr"



# crear interfaz interna al OvS con VLAN ID asignado, asumo que va al unico bridge y servira como gateway de la red VLAN
ovs-vsctl add-port "$brigde" "$NombreRed" tag="$VLAN_ID" -- set interface "$NombreRed" type=internal 
ip link set dev "$NombreRed" up     #  "$bridge"  



# Crear Linux Network Namespace para albergar el servicio DHCP
ip netns add "$NombreRed-dhcp"

# crear interfaces veth 
ip link add "$NombreRed-veth0" type veth peer name "$NombreRed-veth1"

# asignar veth0 al netns dhcp
ip link set "$NombreRed-veth0" netns "$NombreRed-dhcp"
# asignar veth1 al ovs
ovs-vsctl add-port "$brigde" "$NombreRed-veth1" tag="$VLAN_ID"


# prender interfaz veth1 del ovs
ip link set "$NombreRed-veth1" up

# prender interfaces loopback y veth0 del netns dhcp
ip netns exec "$NombreRed-dhcp" ip link set dev lo up
ip netns exec "$NombreRed-dhcp" ip link set dev "$NombreRed-veth0" up

#  prender el bridge ovs
# ip link set dev "$brigde" up


# configurar interfaz interna del ovs correspondiente a la vlan
ip address add "$primera_direccion_disponible_cidr" dev "$NombreRed"    # o "$bridge"  


# asignar segunda direccion ip al servicio dhcp
ip netns exec "$NombreRed-dhcp" ip addr add "$segunda_direccion_disponible_cidr" dev "$NombreRed-veth0"


# Configurar DHCP con dnsmasq

ip netns exec "$NombreRed-dhcp" dnsmasq --interface="$NombreRed-veth0" --dhcp-range=$RangoDHCP --dhcp-option=3,"$primera_direccion_disponible_sincdr" --dhcp-option=6,8.8.8.8,8.8.4.4

# Mostrar información
echo "Red interna $VLAN_ID del orquestador creada correctamente."
sleep 1