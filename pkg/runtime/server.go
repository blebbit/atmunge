package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
	"github.com/bluesky-social/indigo/atproto/crypto"
	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/did"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	plcdb "github.com/blebbit/plc-mirror/pkg/db"
	"github.com/blebbit/plc-mirror/pkg/plc"
)

func (r *Runtime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler(w, req)
}

func (r *Runtime) Ready(w http.ResponseWriter, req *http.Request) {
	convreq.Wrap(func(ctx context.Context) convreq.HttpResponse {
		ts, err := r.LastRecordTimestamp(ctx)
		if err != nil {
			return respond.InternalServerError(err.Error())
		}
		delay := time.Since(ts)
		if delay > r.MaxDelay {
			return respond.ServiceUnavailable(fmt.Sprintf("still %s behind", delay))
		}
		return respond.String("OK")
	})(w, req)
}

func (r *Runtime) Info(w http.ResponseWriter, req *http.Request) {
	convreq.Wrap(func(ctx context.Context) convreq.HttpResponse {
		start := time.Now()
		updateMetrics := func(c int) {
			requestCount.WithLabelValues(fmt.Sprint(c)).Inc()
			requestLatency.WithLabelValues(fmt.Sprint(c)).Observe(float64(time.Since(start)) / float64(time.Millisecond))
		}

		// path component is requested account
		reqAcct := strings.ToLower(strings.TrimPrefix(req.URL.Path, "/info/"))

		// return respond.String(reqAcct)

		var entry plcdb.AccountInfo
		if strings.HasPrefix(reqAcct, "did:") {
			// it looks like did
			err := r.db.Model(&entry).Where("did = ?", reqAcct).Limit(1).Take(&entry).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				updateMetrics(http.StatusNotFound)
				return respond.NotFound("unknown DID")
			}
		} else {
			// otherwise assume a handle
			err := r.db.Model(&entry).Where("handle = ?", reqAcct).Limit(1).Take(&entry).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				updateMetrics(http.StatusNotFound)
				return respond.NotFound("unknown handle")
			}
		}

		view := plcdb.AccountViewFromInfo(&entry)

		updateMetrics(http.StatusOK)
		return respond.JSON(view)
	})(w, req)
}

func (r *Runtime) DidDoc(w http.ResponseWriter, req *http.Request) {
	convreq.Wrap(func(ctx context.Context) convreq.HttpResponse {

		start := time.Now()
		updateMetrics := func(c int) {
			requestCount.WithLabelValues(fmt.Sprint(c)).Inc()
			requestLatency.WithLabelValues(fmt.Sprint(c)).Observe(float64(time.Since(start)) / float64(time.Millisecond))
		}

		// Check if the mirror is up to date.
		ts, err := r.LastRecordTimestamp(ctx)
		if err != nil {
			return respond.InternalServerError(err.Error())
		}
		delay := time.Since(ts)
		if delay > r.MaxDelay {
			// Check LastCompletion and if it's recent enough - that means
			// that we're actually caught up and there simply aren't any recent
			// PLC operations.
			completionDelay := time.Since(r.LastCompletion())
			if completionDelay > r.MaxDelay {
				updateMetrics(http.StatusServiceUnavailable)
				return respond.ServiceUnavailable(fmt.Sprintf("mirror is %s behind", delay))
			}
		}
		log := zerolog.Ctx(ctx)

		requestedDid := strings.ToLower(strings.TrimPrefix(req.URL.Path, "/"))

		// lookup entry in db
		var entry plcdb.PLCLogEntry
		err = r.db.Model(&entry).Where("did = ? AND (NOT nullified)", requestedDid).Order("plc_timestamp desc").Limit(1).Take(&entry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			updateMetrics(http.StatusNotFound)
			return respond.NotFound("unknown DID")
		}
		if err != nil {
			log.Error().Err(err).Str("did", requestedDid).Msgf("Failed to get the last log entry for %q: %s", requestedDid, err)
			updateMetrics(http.StatusInternalServerError)
			return respond.InternalServerError("failed to get the last log entry")
		}

		// check if account deleted
		if _, ok := entry.Operation.Value.(plc.Tombstone); ok {
			updateMetrics(http.StatusNotFound)
			return respond.NotFound("DID deleted")
		}

		// handle legacy
		var op plc.Op
		switch v := entry.Operation.Value.(type) {
		case plc.Op:
			op = v
		case plc.LegacyCreateOp:
			op = v.AsUnsignedOp()
		}

		didValue := did.DID{
			Method: "plc",
			ID:     strings.TrimPrefix(entry.DID, "did:plc:"),
		}
		doc := did.Document{
			Context: []interface{}{
				"https://www.w3.org/ns/did/v1",
				"https://w3id.org/security/multikey/v1"},
			ID:          didValue,
			AlsoKnownAs: mapSlice(op.AlsoKnownAs, ssi.MustParseURI),
		}

		for id, s := range op.Services {
			doc.Service = append(doc.Service, did.Service{
				ID:              ssi.MustParseURI("#" + id),
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
		updateMetrics(http.StatusOK)
		return respond.JSON(doc)

	})(w, req)
}

func mapSlice[A any, B any](s []A, fn func(A) B) []B {
	r := make([]B, 0, len(s))
	for _, v := range s {
		r = append(r, fn(v))
	}
	return r
}
