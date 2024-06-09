package structs

import (
	"time"
)

type ListTemplates struct {
	Result    string `json:"result"`
	Templates []struct {
		TemplateID   string    `json:"template_id"`
		Name         string    `json:"name"`
		CreatedAt    time.Time `json:"created_at"`
		Description  string    `json:"description"`
		TopologyType string    `json:"topology_type"`
	} `json:"templates"`
}

type ListTemplateById struct {
	Result   string `json:"result"`
	Template struct {
		TemplateID       string    `json:"template_id"`
		CreatedAt        time.Time `json:"created_at"`
		AvailabilityZone string    `json:"availability_zone"`
		Deployed         bool      `json:"deployed"`
		Description      string    `json:"description"`
		Name             string    `json:"name"`
		Topology         struct {
			Links []struct {
				LinkID string `json:"link_id"`
				Source string `json:"source"`
				Target string `json:"target"`
			} `json:"links"`
			Nodes []struct {
				NodeID         string  `json:"node_id"`
				Name           string  `json:"name"`
				AccessProtocol string  `json:"access_protocol"`
				CPU            int     `json:"cpu"`
				Image          string  `json:"image"`
				Memory         float32 `json:"memory"`
				SecurityRules  []int   `json:"security_rules"`
				Storage        float32 `json:"storage"`
			} `json:"nodes"`
		} `json:"topology"`
		UserID       string `json:"user_id"`
		VlanID       string `json:"vlan_id"`
		TopologyType string `json:"topology_type"`
	} `json:"template"`
}

type NormalTemplate struct {
	CreatedAt        time.Time `json:"created_at"`
	AvailabilityZone string    `json:"availability_zone"`
	Deployed         bool      `json:"deployed"`
	Description      string    `json:"description"`
	Name             string    `json:"name"`
	Topology         struct {
		Links []struct {
			LinkID string `json:"link_id"`
			Source string `json:"source"`
			Target string `json:"target"`
		} `json:"links"`
		Nodes []struct {
			NodeID         string  `json:"node_id"`
			Name           string  `json:"name"`
			AccessProtocol string  `json:"access_protocol"`
			CPU            int     `json:"cpu"`
			Image          string  `json:"image"`
			Memory         float32 `json:"memory"`
			SecurityRules  []int   `json:"security_rules"`
			Storage        float32 `json:"storage"`
		} `json:"nodes"`
	} `json:"topology"`
	UserID       string `json:"user_id"`
	VlanID       string `json:"vlan_id"`
	TopologyType string `json:"topology_type"`
}
type NormalResponse struct {
	Msg    string `json:"msg"`
	Result string `json:"result"`
}
