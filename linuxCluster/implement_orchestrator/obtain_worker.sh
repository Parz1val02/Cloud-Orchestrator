#!/bin/bash

# Check if the VLAN number is provided as an argument
if [ -z "$1" ]; then
	echo "Usage: $0 <vlan_number>"
	exit 1
fi

# Store the VLAN number in a variable
vlan_number=$1

# Execute the ps aux command, filter qemu processes, and filter based on the VLAN number
ps aux | grep qemu | grep "id=.*vlan${vlan_number}-tap" | while read -r line; do
	# Extract the process ID (second column)
	process_id=$(echo $line | awk '{print $2}')
	# Extract the id parameter and isolate the part after = and before vlan
	id_param=$(echo $line | grep -oP '(?<=-name )\S+' | awk '{print $1}')
	# Extract the VNC port number after :
	vnc_port=$(echo $line | grep -oP "(?<=-vnc 0.0.0.0:)[0-9]+")
	# Calculate the open VNC port
	open_port=$((5900 + vnc_port))
	# Print the isolated id parameter part, the process ID, and the open VNC port
	echo "$id_param $process_id $open_port"
done
