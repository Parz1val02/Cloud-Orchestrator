import requests
from prometheus_client.parser import text_string_to_metric_families
import json

# Fetch metrics from Node Exporter
response = requests.get("http://localhost:9100/metrics")
metrics_data = response.text


# Parse metrics
metrics = {}
for family in text_string_to_metric_families(metrics_data):
    for sample in family.samples:
        if sample.name == 'node_network_receive_bytes_total':
            value = sample.value
            if 'node_network_receive_bytes_total' not in metrics:
                metrics['node_network_receive_bytes_total'] = []
            metrics['node_network_receive_bytes_total'].append(value) 

#node_network_receive_packets_total


# Convert to JSON
metrics_json = json.dumps(metrics)

print(metrics_json)

