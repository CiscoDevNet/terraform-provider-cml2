package cmlclient

// {
// 	"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"state": "DEFINED_ON_CORE",
// 	"created": "2021-02-28T07:33:47+00:00",
// 	"modified": "2021-02-28T07:33:47+00:00",
// 	"lab_title": "Lab at Mon 17:27 PM",
// 	"owner": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 	"lab_description": "string",
// 	"node_count": 0,
// 	"link_count": 0,
// 	"lab_notes": "string",
// 	"groups": [
// 	  {
// 		"id": "90f84e38-a71c-4d57-8d90-00fa8a197385",
// 		"permission": "read_only"
// 	  }
// 	]
// }

const (
	LabStateDefined = "DEFINED_ON_CORE"
	LabStateStopped = "STOPPED"
	LabStateStarted = "STARTED"
)

type IDlist []string
type nodeMap map[string]*Node
type interfaceMap map[string]*Interface
type linkList []*Link

type labAlias struct {
	Lab
	OwnerID string `json:"owner"`
}

type Lab struct {
	ID          string   `json:"id"`
	State       string   `json:"state"`
	Created     string   `json:"created"`
	Modified    string   `json:"modified"`
	Title       string   `json:"lab_title"`
	Description string   `json:"lab_description"`
	Notes       string   `json:"lab_notes"`
	Owner       *User    `json:"owner"`
	NodeCount   int      `json:"node_count"`
	LinkCount   int      `json:"link_count"`
	Nodes       nodeMap  `json:"nodes"`
	Links       linkList `json:"links"`
}

func (l *Lab) CanBeWiped() bool {
	if len(l.Nodes) == 0 {
		return l.State != LabStateDefined
	}
	for _, node := range l.Nodes {
		if node.State != LabStateDefined {
			return true
		}
	}
	return false
}

func (l *Lab) Running() bool {
	for _, node := range l.Nodes {
		if node.State != LabStateDefined {
			return true
		}
	}
	return false
}

type LabImport struct {
	ID       string   `json:"id"`
	Warnings []string `json:"warnings"`
}
