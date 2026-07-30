// Package discoveryapi contains the shared HTTP projection for serving an
// instance-local Foundation Discovery publication to Nuxt transports.
package discoveryapi

import "github.com/yueli-official/foundation/go/discovery"

type ArtifactView struct {
	Name      string `json:"name"`
	MediaType string `json:"mediaType"`
	Content   string `json:"content"`
	Bytes     int64  `json:"bytes"`
	SHA256    string `json:"sha256"`
}

type ArtifactResult struct {
	ContractVersion string        `json:"contractVersion"`
	Artifact        *ArtifactView `json:"artifact"`
}

func FindArtifact(snapshot discovery.MemoryPublication, name string) ArtifactResult {
	result := ArtifactResult{
		ContractVersion: snapshot.Manifest.ContractVersion,
	}
	for _, artifact := range snapshot.Manifest.Artifacts {
		if artifact.Name != name {
			continue
		}
		content, ok := snapshot.Artifacts[artifact.Name]
		if !ok {
			return result
		}
		result.Artifact = &ArtifactView{
			Name: artifact.Name, MediaType: artifact.MediaType,
			Content: string(content), Bytes: artifact.Bytes, SHA256: artifact.SHA256,
		}
		return result
	}
	return result
}
