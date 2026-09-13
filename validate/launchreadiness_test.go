package validate

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func testLaunchArchitecture(rel sas.Relationship, targetOwner string) *sas.Architecture {
	return &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "user", Kind: sas.NodeKindActor, Name: "User"},
			{ID: "web", Kind: sas.NodeKindService, Name: "Web", Owner: targetOwner, Boundaries: []string{"trust-app"}},
		},
		Boundaries: []sas.Boundary{
			{ID: "trust-app", Kind: sas.BoundaryKindTrust, Name: "App Trust Zone"},
		},
		Relationships: []sas.Relationship{rel},
	}
}

func TestIsInternetFacing_ActorIntoBoundary(t *testing.T) {
	arch := testLaunchArchitecture(sas.Relationship{ID: "user-to-web", From: "user", To: "web"}, "team")
	if !isInternetFacing(arch, arch.Relationships[0]) {
		t.Error("expected actor calling into a boundary to be internet-facing")
	}
}

func TestIsInternetFacing_InternalCallNotFacing(t *testing.T) {
	arch := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "web", Kind: sas.NodeKindService, Boundaries: []string{"trust-app"}},
			{ID: "api", Kind: sas.NodeKindService, Boundaries: []string{"trust-app"}},
		},
		Boundaries:    []sas.Boundary{{ID: "trust-app", Kind: sas.BoundaryKindTrust}},
		Relationships: []sas.Relationship{{ID: "web-to-api", From: "web", To: "api"}},
	}
	if isInternetFacing(arch, arch.Relationships[0]) {
		t.Error("did not expect an internal service-to-service call to be internet-facing")
	}
}

func TestIsInternetFacing_OutboundToExternalNotFacing(t *testing.T) {
	// Our own service calling out to an external OAuth provider is the
	// opposite direction from a public endpoint receiving traffic.
	arch := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Boundaries: []string{"trust-app"}},
			{ID: "oauth", Kind: sas.NodeKindExternalService},
		},
		Boundaries:    []sas.Boundary{{ID: "trust-app", Kind: sas.BoundaryKindTrust}},
		Relationships: []sas.Relationship{{ID: "api-to-oauth", From: "api", To: "oauth"}},
	}
	if isInternetFacing(arch, arch.Relationships[0]) {
		t.Error("did not expect our own service calling out to be internet-facing")
	}
}

func TestIsInternetFacing_NoBoundaryCrossingNotFacing(t *testing.T) {
	// An actor calling a node with no boundary membership crosses
	// nothing, so there is no boundary being "entered."
	arch := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "user", Kind: sas.NodeKindActor},
			{ID: "web", Kind: sas.NodeKindService},
		},
		Relationships: []sas.Relationship{{ID: "user-to-web", From: "user", To: "web"}},
	}
	if isInternetFacing(arch, arch.Relationships[0]) {
		t.Error("did not expect a boundary-less relationship to be internet-facing")
	}
}

func TestValidate_LaunchReadiness_AllGapsReported(t *testing.T) {
	arch := testLaunchArchitecture(sas.Relationship{ID: "user-to-web", From: "user", To: "web"}, "") // no owner, no transport, no identity, no data

	findings := Validate(arch, ProfileSecurity)
	for _, want := range []string{
		"launch-readiness.internet-facing-tls-required",
		"launch-readiness.internet-facing-authn-required",
		"launch-readiness.internet-facing-data-classification-required",
		"launch-readiness.internet-facing-owner-required",
	} {
		if !hasRule(findings, want) {
			t.Errorf("expected finding %s, got %+v", want, findings)
		}
	}
}

func TestValidate_LaunchReadiness_Passes(t *testing.T) {
	arch := testLaunchArchitecture(sas.Relationship{
		ID: "user-to-web", From: "user", To: "web",
		Transport: &sas.Transport{Encryption: "tls"},
		Identity:  &sas.IdentityRef{NodeID: "user"},
		Data:      &sas.DataFlow{Classifications: []string{"session_token"}},
	}, "web-team")

	findings := Validate(arch, ProfileSecurity)
	for _, f := range findings {
		if f.RuleID == "launch-readiness.internet-facing-tls-required" ||
			f.RuleID == "launch-readiness.internet-facing-authn-required" ||
			f.RuleID == "launch-readiness.internet-facing-data-classification-required" ||
			f.RuleID == "launch-readiness.internet-facing-owner-required" {
			t.Errorf("did not expect launch-readiness finding, got %+v", f)
		}
	}
}

func TestValidate_LaunchReadiness_InternalCallExempt(t *testing.T) {
	arch := &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "web", Kind: sas.NodeKindService, Boundaries: []string{"trust-app"}},
			{ID: "api", Kind: sas.NodeKindService, Boundaries: []string{"trust-app"}},
		},
		Boundaries:    []sas.Boundary{{ID: "trust-app", Kind: sas.BoundaryKindTrust}},
		Relationships: []sas.Relationship{{ID: "web-to-api", From: "web", To: "api"}},
	}

	findings := Validate(arch, ProfileSecurity)
	for _, f := range findings {
		if f.Profile == ProfileSecurity && f.RuleID != "" {
			if f.RuleID == "launch-readiness.internet-facing-tls-required" {
				t.Errorf("did not expect launch-readiness finding for an internal call, got %+v", f)
			}
		}
	}
}
