package grapher

import (
	"context"
	"io"

	"bitbucket.org/atlassian/go-vpcflow"
	"bitbucket.org/atlassian/vpcflow-grapherd/pkg/types"
)

// DOT is a grapher module which converts a VPC flow log digest into a DOT graph using the go-vpc library.
// If successful, it stores the resulting graph in the backend implemented by the provided types.Storage
type DOT struct {
	Converter vpcflow.Converter
	Storage   types.Storage
}

// Graph graphs the given digest in DOT format, and stores the generated DOT contents identified by the supplied id
func (g *DOT) Graph(ctx context.Context, id string, digest io.ReadCloser) error {
	r, err := g.Converter(digest)
	if err != nil {
		return err
	}
	defer r.Close()
	return g.Storage.Store(ctx, id, r)
}
