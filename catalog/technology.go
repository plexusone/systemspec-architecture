package catalog

import "github.com/plexusone/systemspec-architecture/sas"

// providerCatalogs maps a Technology.Provider to its display-name lookup
// table. Adding a new provider catalog means adding one entry here plus
// its own <provider>.go file — the core sas package never changes.
var providerCatalogs = map[string]map[string]string{
	"aws": awsDisplayNames,
	"gcp": gcpDisplayNames,
	"k8s": k8sDisplayNames,
}

// DisplayName returns a human-friendly display name for a Technology,
// e.g. {provider: aws, service: lambda} -> "AWS Lambda". Falls back to
// "<Provider>/<Service>" (or just Provider when Service is empty) when
// the combination is not in the catalog, and to "" when tech is nil.
func DisplayName(tech *sas.Technology) string {
	if tech == nil {
		return ""
	}
	if names, ok := providerCatalogs[tech.Provider]; ok {
		if name, ok := names[tech.Service]; ok {
			return name
		}
	}
	if tech.Service == "" {
		return tech.Provider
	}
	return tech.Provider + "/" + tech.Service
}

// IconHint returns a short, catalog-namespaced identifier suitable for
// icon lookups by future graphical renderers, e.g. "aws-lambda". Falls
// back to "<provider>-<service>", and to "" when tech is nil.
func IconHint(tech *sas.Technology) string {
	if tech == nil {
		return ""
	}
	if tech.Service == "" {
		return tech.Provider
	}
	return tech.Provider + "-" + tech.Service
}
