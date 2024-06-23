#!/bin/bash

# parámetros
VLAN_ID="$1"  # ID de la VLAN para la cual se proporcionará acceso a Internet
InternetInterface="ens3"  # Interfaz de red conectada a Internet en el HeadNode

# dirección IP de la interfaz de Internet
# InternetIP=$(ip addr show dev $InternetInterface | grep -oP '(?<=inet\s)\d+(\.\d+){3}')




# Obtener el nombre de la interfaz asociada con la VLAN
interface_name=$(sudo ovs-vsctl list port | grep "vlan$VLAN_ID" | awk '{print $3}' | tail -n 1)
# Obtener la información de la dirección IP y la máscara de subred de la interfaz
interface_info=$(ip addr show dev $interface_name)
if [ -z "$interface_info" ]; then
    echo "No se encontró ninguna interfaz con VLAN ID $VLAN_ID"
    exit 1
fi

echo "interface info : $interface_info"
# Extraer la dirección de red y la máscara de subred de la salida
# ip_address=$(echo "$interface_info" | awk '/inet / {print $2}') # extrae todo en formato cidr
ip_address=$(echo "$interface_info" | awk '/inet / {split($2, ip_parts, "/"); print ip_parts[1]}')
# Extraer la dirección IP y la máscara de subred de la salida
subnet_mask=$(echo "$interface_info" | awk '/inet / {print $2}' | awk -F/ '{print $2}')
# Calcular la dirección de red a partir de la dirección IP y la máscara de subred
# ip_network=$(ipcalc -n $ip_address/$subnet_mask | awk '/Network/ {print $2}') # extrae la direcc de red peroo con formato cidr
ip_network=$(ipcalc -n $ip_address/$subnet_mask | awk '/Network/ {split($2, parts, "/"); print parts[1]}')
# Imprimir la dirección de red en formato CIDR
echo "Dirección de red asociada a la VLAN ID $VLAN_ID (formato CIDR): $ip_network/$subnet_mask"
vlan_cidr="$ip_network/$subnet_mask"
echo $vlan_cidr


# regla de iptables para habilitar el enrutamiento
iptables -I FORWARD -s $vlan_cidr -j ACCEPT
iptables -I FORWARD -d $vlan_cidr -j ACCEPT
iptables -t nat -A POSTROUTING -s $vlan_cidr -j MASQUERADE

# Habilitar el enrutamiento IPv4
echo 1 > /proc/sys/net/ipv4/ip_forward

# Asegurarse de que el tráfico de retorno esté permitido
# iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

echo "Acceso a Internet habilitado para VLAN $VLAN_ID con dirección de red $vlan_cidr"
