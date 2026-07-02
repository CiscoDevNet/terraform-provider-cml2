package lifecycle

import (
	"testing"

	"github.com/rschmied/gocmlclient/pkg/models"
)

func TestLabHasDrift(t *testing.T) {
	tests := []struct {
		name    string
		lab     models.Lab
		desired models.LabState
		want    bool
	}{
		{
			name:    "started no drift (booted nodes + started links)",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateBooted},
					models.UUID("n2"): {ID: models.UUID("n2"), State: models.NodeStateStarted},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: false,
		},
		{
			name:    "started drift on node stopped",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStopped},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: true,
		},
		{
			name:    "started drift on link stopped",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStarted},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStopped}},
			},
			want: true,
		},
		{
			name:    "stopped no drift",
			desired: models.LabStateStopped,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStopped},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStopped}},
			},
			want: false,
		},
		{
			name:    "stopped drift on defined node",
			desired: models.LabStateStopped,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateDefined},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStopped}},
			},
			want: true,
		},
		{
			name:    "defined no drift",
			desired: models.LabStateDefined,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateDefined},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateDefined}},
			},
			want: false,
		},
		{
			name:    "defined drift on started link",
			desired: models.LabStateDefined,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateDefined},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: true,
		},
		{
			name:    "nil node pointers are ignored",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): nil,
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labHasDrift(&tt.lab, tt.desired)
			if got != tt.want {
				t.Fatalf("labHasDrift() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLabHasDrift_Edges(t *testing.T) {
	t.Run("nil lab panics (document current behavior)", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatalf("expected panic for nil lab")
			}
		}()
		_ = labHasDrift(nil, models.LabStateStarted)
	})

	tests := []struct {
		name    string
		lab     models.Lab
		desired models.LabState
		want    bool
	}{
		{
			name:    "unknown desired state returns no drift",
			desired: models.LabState("BOGUS"),
			lab: models.Lab{
				Nodes: models.NodeMap{models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStarted}},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: false,
		},
		{
			name:    "empty lab returns no drift for started",
			desired: models.LabStateStarted,
			lab:     models.Lab{},
			want:    false,
		},
		{
			name:    "empty lab returns no drift for stopped",
			desired: models.LabStateStopped,
			lab:     models.Lab{},
			want:    false,
		},
		{
			name:    "empty lab returns no drift for defined",
			desired: models.LabStateDefined,
			lab:     models.Lab{},
			want:    false,
		},
		{
			name:    "started treats queued as drift",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateQueued}},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: true,
		},
		{
			name:    "started treats disconnected as drift",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateDisconnected}},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: true,
		},
		{
			name:    "stopped treats defined node as drift (differs from modify_plan tolerance)",
			desired: models.LabStateStopped,
			lab: models.Lab{
				Nodes: models.NodeMap{models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateDefined}},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStopped}},
			},
			want: true,
		},
		{
			name:    "mixed states any bad node yields drift",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{
					models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStarted},
					models.UUID("n2"): {ID: models.UUID("n2"), State: models.NodeStateStopped},
				},
				Links: models.LinkList{{ID: models.UUID("l1"), State: models.LinkStateStarted}},
			},
			want: true,
		},
		{
			name:    "mixed states any bad link yields drift",
			desired: models.LabStateStarted,
			lab: models.Lab{
				Nodes: models.NodeMap{models.UUID("n1"): {ID: models.UUID("n1"), State: models.NodeStateStarted}},
				Links: models.LinkList{
					{ID: models.UUID("l1"), State: models.LinkStateStarted},
					{ID: models.UUID("l2"), State: models.LinkStateStopped},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labHasDrift(&tt.lab, tt.desired)
			if got != tt.want {
				t.Fatalf("labHasDrift() = %v, want %v", got, tt.want)
			}
		})
	}
}
