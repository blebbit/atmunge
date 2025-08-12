package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// ResolveDid resolves a handle or DID to a DID and a PDS endpoint.
func (r *Runtime) ResolveDid(ctx context.Context, handleOrDID string) (string, string, error) {
	var did syntax.DID
	var err error

	if strings.HasPrefix(handleOrDID, "did:") {
		did, err = syntax.ParseDID(handleOrDID)
		if err != nil {
			return "", "", fmt.Errorf("invalid did: %w", err)
		}
	} else {
		d, err := r.resolveHandle(ctx, handleOrDID)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve handle: %w", err)
		}
		did = d
	}

	pds, err := r.lookupPDS(ctx, did)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup pds: %w", err)
	}

	return did.String(), pds, nil
}

func (r *Runtime) resolveHandle(ctx context.Context, handle string) (syntax.DID, error) {
	h, err := syntax.ParseHandle(handle)
	if err != nil {
		return "", fmt.Errorf("invalid handle: %w", err)
	}
	dir := identity.DefaultDirectory()
	ident, err := dir.LookupHandle(ctx, h)
	if err != nil {
		return "", err
	}
	return ident.DID, nil
}

func (r *Runtime) lookupPDS(ctx context.Context, did syntax.DID) (string, error) {
	dir := identity.DefaultDirectory()
	ident, err := dir.LookupDID(ctx, did)
	if err != nil {
		return "", err
	}
	if ident.PDSEndpoint() == "" {
		return "", fmt.Errorf("no PDS endpoint found for %s", did)
	}
	return ident.PDSEndpoint(), nil
}
