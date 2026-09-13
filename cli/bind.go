package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/plexusone/systemspec-architecture/sas"
)

// pidlProtocol is the minimal subset of a PIDL document binding
// verification needs: which entity IDs the protocol declares. SAS does
// not depend on PIDL's Go module — this is a narrow, local decode of the
// fields that matter here, not a PIDL implementation.
type pidlProtocol struct {
	Protocol struct {
		ID string `json:"id"`
	} `json:"protocol"`
	Entities []struct {
		ID string `json:"id"`
	} `json:"entities"`
}

// BindOptions configures a Bind run.
type BindOptions struct {
	// ArchitecturePath is the path to a SAS architecture JSON document.
	// Required.
	ArchitecturePath string

	// PIDLPath is the path to a PIDL protocol JSON document. Required.
	PIDLPath string
}

// BindingCheck is the result of checking one ProtocolBinding against the
// PIDL protocol it references.
type BindingCheck struct {
	BindingID string `json:"bindingId"`

	// UnknownEntities lists Participants keys that do not match any
	// entity ID declared in the PIDL protocol.
	UnknownEntities []string `json:"unknownEntities,omitempty"`

	// UnresolvedNodes lists Participants values that do not resolve to a
	// node in the architecture.
	UnresolvedNodes []string `json:"unresolvedNodes,omitempty"`

	// UnboundEntities lists PIDL entity IDs the protocol declares that
	// this binding does not assign a participant for. This is
	// informational, not an error: not every architecture needs a node
	// for every protocol entity (e.g. an ephemeral browser redirect).
	UnboundEntities []string `json:"unboundEntities,omitempty"`
}

// OK reports whether this binding has no unknown entities or unresolved
// nodes. Unbound entities alone do not fail a binding.
func (c BindingCheck) OK() bool {
	return len(c.UnknownEntities) == 0 && len(c.UnresolvedNodes) == 0
}

// BindResult is the outcome of a Bind run.
type BindResult struct {
	ProtocolID string         `json:"protocolId"`
	Bindings   []BindingCheck `json:"bindings"`
}

// OK reports whether every binding in the result is OK. A result with no
// bindings at all (nothing in the architecture references this protocol)
// is vacuously OK — that is a fact worth reporting, not a failure.
func (r BindResult) OK() bool {
	for _, b := range r.Bindings {
		if !b.OK() {
			return false
		}
	}
	return true
}

// Bind loads the architecture at opts.ArchitecturePath and the PIDL
// protocol at opts.PIDLPath, and checks every ProtocolBinding in the
// architecture whose ProtocolRef is "pidl://<protocol.id>" against the
// protocol's declared entities: every participant key must be a real
// entity ID, every participant value must resolve to a real node.
func Bind(opts BindOptions) (BindResult, error) {
	archData, err := os.ReadFile(opts.ArchitecturePath)
	if err != nil {
		return BindResult{}, fmt.Errorf("read %s: %w", opts.ArchitecturePath, err)
	}
	var arch sas.Architecture
	if err := json.Unmarshal(archData, &arch); err != nil {
		return BindResult{}, fmt.Errorf("parse %s: %w", opts.ArchitecturePath, err)
	}

	pidlData, err := os.ReadFile(opts.PIDLPath)
	if err != nil {
		return BindResult{}, fmt.Errorf("read %s: %w", opts.PIDLPath, err)
	}
	var protocol pidlProtocol
	if err := json.Unmarshal(pidlData, &protocol); err != nil {
		return BindResult{}, fmt.Errorf("parse %s: %w", opts.PIDLPath, err)
	}
	if protocol.Protocol.ID == "" {
		return BindResult{}, fmt.Errorf("%s has no protocol.id", opts.PIDLPath)
	}

	entityIDs := make(map[string]bool, len(protocol.Entities))
	for _, e := range protocol.Entities {
		entityIDs[e.ID] = true
	}

	protocolRef := "pidl://" + protocol.Protocol.ID
	result := BindResult{ProtocolID: protocol.Protocol.ID}

	for _, binding := range arch.BindingsForProtocol(protocolRef) {
		check := BindingCheck{BindingID: binding.ID}
		bound := make(map[string]bool, len(binding.Participants))

		for entityID, nodeID := range binding.Participants {
			bound[entityID] = true
			if !entityIDs[entityID] {
				check.UnknownEntities = append(check.UnknownEntities, entityID)
			}
			if _, ok := arch.NodeByID(nodeID); !ok {
				check.UnresolvedNodes = append(check.UnresolvedNodes, nodeID)
			}
		}
		for _, e := range protocol.Entities {
			if !bound[e.ID] {
				check.UnboundEntities = append(check.UnboundEntities, e.ID)
			}
		}

		sort.Strings(check.UnknownEntities)
		sort.Strings(check.UnresolvedNodes)
		sort.Strings(check.UnboundEntities)

		result.Bindings = append(result.Bindings, check)
	}

	return result, nil
}
