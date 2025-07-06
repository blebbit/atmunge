package plc

import (
	"slices"
	"strings"

	"github.com/bluesky-social/indigo/atproto/crypto"
	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/did"
)

func MakeDoc(entry OperationLogEntry, op Op) (did.Document, error) {
	didValue := did.DID{
		Method: "plc",
		ID:     strings.TrimPrefix(entry.DID, "did:plc:"),
	}
	aka, err := mapSlice(op.AlsoKnownAs, ssi.ParseURI)
	if err != nil {
		return did.Document{}, err
	}
	doc := did.Document{
		Context: []interface{}{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/multikey/v1"},
		ID:          didValue,
		AlsoKnownAs: aka,
	}

	for id, s := range op.Services {
		sid, err := ssi.ParseURI("#" + id)
		if err != nil {
			return did.Document{}, err
		}
		doc.Service = append(doc.Service, did.Service{
			ID:              *sid,
			Type:            s.Type,
			ServiceEndpoint: s.Endpoint,
		})
	}

	for id, m := range op.VerificationMethods {
		idValue := did.DIDURL{
			DID:      didValue,
			Fragment: id,
		}
		doc.VerificationMethod.Add(&did.VerificationMethod{
			ID:                 idValue,
			Type:               "Multikey",
			Controller:         didValue,
			PublicKeyMultibase: strings.TrimPrefix(m, "did:key:"),
		})

		key, err := crypto.ParsePublicDIDKey(m)
		if err == nil {
			context := ""
			switch key.(type) {
			case *crypto.PublicKeyK256:
				context = "https://w3id.org/security/suites/secp256k1-2019/v1"
			case *crypto.PublicKeyP256:
				context = "https://w3id.org/security/suites/ecdsa-2019/v1"
			}
			if context != "" && !slices.Contains(doc.Context, interface{}(context)) {
				doc.Context = append(doc.Context, context)
			}
		}
	}

	return doc, nil
}

func mapSlice[A any, B any](s []A, fn func(A) (*B, error)) ([]B, error) {
	var e error
	r := make([]B, 0, len(s))
	for _, v := range s {
		n, err := fn(v)
		if err != nil {
			e = err
			continue
		}
		r = append(r, *n)
	}
	return r, e
}
