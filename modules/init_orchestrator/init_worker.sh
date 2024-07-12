#!/bin/bash

# parámetros
nombreOvS="$1"
InterfacesAConectar="$2"

# crear ovs local al worker# si no existe
if ! ovs-vsctl br-exists "$nombreOvS"; then
	ovs-vsctl add-br "$nombreOvS"
	echo "Open vSwitch bridge '$nombreOvS' created."
else
	echo "Open vSwitch bridge '$nombreOvS' already exists."
fi

# conectar lista de interfaces al ovs (ens4 salida del worker hacia OFS y
# deben crearse tantas interfaces TAP para vincularse con las tantas VMs a crear en el worker)
for iface in $InterfacesAConectar; do
	ovs-vsctl add-port "$nombreOvS" "$iface"
done

ip link set dev "$nombreOvS" up

# mostrar confirmacion de inicializacion
echo "Worker inicializado correctamente."
