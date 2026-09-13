package catalog

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func TestDisplayName_KnownCombinations(t *testing.T) {
	tests := []struct {
		tech sas.Technology
		want string
	}{
		{sas.Technology{Provider: "aws", Service: "lambda"}, "AWS Lambda"},
		{sas.Technology{Provider: "aws", Service: "rds"}, "Amazon RDS"},
		{sas.Technology{Provider: "gcp", Service: "cloud-run"}, "Google Cloud Run"},
		{sas.Technology{Provider: "k8s", Service: "deployment"}, "Kubernetes Deployment"},
		{sas.Technology{Provider: "k8s", Service: "namespace"}, "Kubernetes Namespace"},
		{sas.Technology{Provider: "k8s", Service: "cluster"}, "Kubernetes Cluster"},
		{sas.Technology{Provider: "k8s", Service: "pod"}, "Kubernetes Pod"},
		{sas.Technology{Provider: "k8s", Service: "service_account"}, "Kubernetes ServiceAccount"},
		{sas.Technology{Provider: "k8s", Service: "rolebinding"}, "Kubernetes RoleBinding"},
	}
	for _, tt := range tests {
		if got := DisplayName(&tt.tech); got != tt.want {
			t.Errorf("DisplayName(%+v) = %q, want %q", tt.tech, got, tt.want)
		}
	}
}

func TestDisplayName_UnknownFallsBackToProviderService(t *testing.T) {
	tech := sas.Technology{Provider: "aws", Service: "some-new-service"}
	if got := DisplayName(&tech); got != "aws/some-new-service" {
		t.Errorf("DisplayName(%+v) = %q, want %q", tech, got, "aws/some-new-service")
	}
}

func TestDisplayName_UnknownProviderNoService(t *testing.T) {
	tech := sas.Technology{Provider: "postgresql"}
	if got := DisplayName(&tech); got != "postgresql" {
		t.Errorf("DisplayName(%+v) = %q, want %q", tech, got, "postgresql")
	}
}

func TestDisplayName_Nil(t *testing.T) {
	if got := DisplayName(nil); got != "" {
		t.Errorf("DisplayName(nil) = %q, want empty string", got)
	}
}

func TestIconHint(t *testing.T) {
	tech := sas.Technology{Provider: "aws", Service: "lambda"}
	if got := IconHint(&tech); got != "aws-lambda" {
		t.Errorf("IconHint(%+v) = %q, want %q", tech, got, "aws-lambda")
	}

	noService := sas.Technology{Provider: "postgresql"}
	if got := IconHint(&noService); got != "postgresql" {
		t.Errorf("IconHint(%+v) = %q, want %q", noService, got, "postgresql")
	}

	if got := IconHint(nil); got != "" {
		t.Errorf("IconHint(nil) = %q, want empty string", got)
	}
}
