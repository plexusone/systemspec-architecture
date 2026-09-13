package validate

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func hasRule(findings []Finding, ruleID string) bool {
	for _, f := range findings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

func TestValidate_CoreReferentialIntegrity(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes:   []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: "A", Boundaries: []string{"missing-boundary"}}},
		Relationships: []sas.Relationship{
			{ID: "a-to-b", From: "a", To: "missing-node", CrossesBoundaries: []string{"missing-boundary-2"}},
		},
	}

	findings := Validate(arch)

	for _, want := range []string{
		"core.node-boundary-resolves",
		"core.relationship-endpoints-resolve",
		"core.relationship-crosses-resolves",
	} {
		if !hasRule(findings, want) {
			t.Errorf("expected finding %s, got %+v", want, findings)
		}
	}
}

func TestValidate_CleanArchitectureNoFindings(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A"},
			{ID: "b", Kind: sas.NodeKindService, Name: "B"},
		},
		Relationships: []sas.Relationship{{ID: "a-to-b", From: "a", To: "b"}},
	}

	findings := Validate(arch)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %+v", findings)
	}
}

func TestValidate_UnknownProfile(t *testing.T) {
	arch := &sas.Architecture{Version: "0.1"}
	findings := Validate(arch, Profile("bogus"))
	if !hasRule(findings, "core.unknown-profile") {
		t.Errorf("expected core.unknown-profile finding, got %+v", findings)
	}
}

func TestValidate_Development_NoExtraRules(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes:   []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: "A"}},
	}
	findings := Validate(arch, ProfileDevelopment)
	if len(findings) != 0 {
		t.Errorf("expected no findings under development profile, got %+v", findings)
	}
}

func TestValidate_Deployment(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A"}, // no technology, no boundary
		},
		Relationships: []sas.Relationship{
			{ID: "a-to-a", From: "a", To: "a"}, // no transport
		},
	}

	findings := Validate(arch, ProfileDeployment)
	for _, want := range []string{
		"deployment.node-technology-required",
		"deployment.node-boundary-required",
		"deployment.relationship-transport-required",
	} {
		if !hasRule(findings, want) {
			t.Errorf("expected finding %s, got %+v", want, findings)
		}
	}
}

func TestValidate_Deployment_Passes(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{
				ID: "a", Kind: sas.NodeKindService, Name: "A",
				Technology: &sas.Technology{Provider: "aws", Service: "ecs"},
				Boundaries: []string{"vpc"},
			},
		},
		Boundaries: []sas.Boundary{{ID: "vpc", Kind: sas.BoundaryKindNetwork, Name: "VPC"}},
	}

	findings := Validate(arch, ProfileDeployment)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %+v", findings)
	}
}

func TestValidate_Security_BoundaryCrossing(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A", Boundaries: []string{"trust-a"}},
			{ID: "b", Kind: sas.NodeKindService, Name: "B", Boundaries: []string{"trust-b"}},
		},
		Boundaries: []sas.Boundary{
			{ID: "trust-a", Kind: sas.BoundaryKindTrust, Name: "A"},
			{ID: "trust-b", Kind: sas.BoundaryKindTrust, Name: "B"},
		},
		Relationships: []sas.Relationship{{ID: "a-to-b", From: "a", To: "b"}}, // no transport, no identity
	}

	findings := Validate(arch, ProfileSecurity)
	for _, want := range []string{
		"security.boundary-crossing-encryption-required",
		"security.boundary-crossing-identity-required",
	} {
		if !hasRule(findings, want) {
			t.Errorf("expected finding %s, got %+v", want, findings)
		}
	}
}

func TestValidate_Security_BoundaryCrossing_Passes(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A", Boundaries: []string{"trust-a"}},
			{ID: "b", Kind: sas.NodeKindService, Name: "B", Boundaries: []string{"trust-b"}},
		},
		Boundaries: []sas.Boundary{
			{ID: "trust-a", Kind: sas.BoundaryKindTrust, Name: "A"},
			{ID: "trust-b", Kind: sas.BoundaryKindTrust, Name: "B"},
		},
		Relationships: []sas.Relationship{
			{
				ID: "a-to-b", From: "a", To: "b",
				Transport: &sas.Transport{Encryption: "tls"},
				Identity:  &sas.IdentityRef{NodeID: "a"},
			},
		},
	}

	findings := Validate(arch, ProfileSecurity)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %+v", findings)
	}
}

func TestValidate_Security_WriteEntitlements(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A"},
			{ID: "b", Kind: sas.NodeKindDataDatabase, Name: "B"},
		},
		Relationships: []sas.Relationship{
			{ID: "a-to-b", From: "a", To: "b", Operations: []sas.Operation{sas.OperationCreate}},
		},
	}

	findings := Validate(arch, ProfileSecurity)
	if !hasRule(findings, "security.write-entitlements-required") {
		t.Errorf("expected security.write-entitlements-required, got %+v", findings)
	}
}

func TestValidate_ThreatModel_ExternalEdge(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A"},
			{ID: "ext", Kind: sas.NodeKindExternalService, Name: "External"},
		},
		Relationships: []sas.Relationship{{ID: "a-to-ext", From: "a", To: "ext"}},
	}

	findings := Validate(arch, ProfileThreatModel)
	if !hasRule(findings, "threat-model.external-edge-data-classification-required") {
		t.Errorf("expected threat-model.external-edge-data-classification-required, got %+v", findings)
	}
}

func TestValidate_ThreatModel_ExternalEdge_Passes(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "a", Kind: sas.NodeKindService, Name: "A"},
			{ID: "ext", Kind: sas.NodeKindExternalService, Name: "External"},
		},
		Relationships: []sas.Relationship{
			{ID: "a-to-ext", From: "a", To: "ext", Data: &sas.DataFlow{Classifications: []string{"customer_data"}}},
		},
	}

	findings := Validate(arch, ProfileThreatModel)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %+v", findings)
	}
}

func TestValidate_SRE_CriticalNode(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes:   []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: "A", Criticality: sas.CriticalityTier0}},
	}

	findings := Validate(arch, ProfileSRE)
	for _, want := range []string{
		"sre.critical-owner-required",
		"sre.critical-availability-target-required",
		"sre.critical-assurance-metrics-required",
	} {
		if !hasRule(findings, want) {
			t.Errorf("expected finding %s, got %+v", want, findings)
		}
	}
}

func TestValidate_SRE_CriticalNode_Passes(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{
				ID: "a", Kind: sas.NodeKindService, Name: "A",
				Criticality: sas.CriticalityTier1,
				Owner:       "platform-team",
				Extensions:  &sas.Extensions{SRE: &sas.SREExtension{AvailabilityTarget: "99.9%"}},
				Assurance:   &sas.Assurance{Metrics: []string{"otel://service/a"}},
			},
		},
	}

	findings := Validate(arch, ProfileSRE)
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %+v", findings)
	}
}

func TestValidate_SRE_NonCriticalNodeExempt(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes:   []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: "A", Criticality: sas.CriticalityTier2}},
	}

	findings := Validate(arch, ProfileSRE)
	if len(findings) != 0 {
		t.Errorf("expected tier2 node to be exempt, got %+v", findings)
	}
}

func TestValidate_MultipleProfilesCombine(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes:   []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: "A"}},
	}

	findings := Validate(arch, ProfileDeployment, ProfileSRE)
	if !hasRule(findings, "deployment.node-technology-required") {
		t.Error("expected deployment rule to run")
	}
	// Node has no criticality, so sre rules are exempt — this proves both
	// profiles ran (no unknown-profile finding) without asserting an sre
	// finding that wouldn't apply here.
	if hasRule(findings, "core.unknown-profile") {
		t.Error("did not expect unknown-profile finding")
	}
}
