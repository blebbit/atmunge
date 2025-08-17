package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/blebbit/at-mirror/pkg/db"
	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/plc"
	"github.com/blebbit/at-mirror/pkg/runtime"
)

type Server struct {
	e *echo.Echo
	r *runtime.Runtime
}

func NewServer(r *runtime.Runtime) *Server {
	e := echo.New()
	s := &Server{
		e: e,
		r: r,
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from AT Mirror!")
	})

	e.GET("/ready", s.Ready)
	e.GET("/:did", s.DidDoc)
	e.GET("/info/:acct", s.Info)
	e.GET("/autocomplete/:token", s.Autocomplete)

	// TODO, endpoints for
	// 1. getting info for multiple accounts
	// 2. syncing an account (maybe private or apikey based to start with, make paid?)
	// enable some apikey based support for people who want theirs public and protected

	return s
}

const readyMsgFmt = `plc:  %s
acct: %s
`

func (s *Server) Echo() *echo.Echo {
	return s.e
}

func (s *Server) Ready(c echo.Context) error {
	var (
		plcStatus  string
		acctStatus string
	)

	// PLC status
	plcStatus = "OK"
	ts, err := s.r.LastRecordTimestamp(s.r.Ctx)
	if err != nil {
		plcStatus = err.Error()
	}
	delay := time.Since(ts)
	if delay > s.r.MaxDelay {
		plcStatus = fmt.Sprintf("still %s behind", delay)
	}

	// return our status message
	return c.String(http.StatusOK, fmt.Sprintf(readyMsgFmt, plcStatus, acctStatus))
}

func (s *Server) Info(c echo.Context) error {
	start := time.Now()
	updateMetrics := func(n int) {
		requestCount.WithLabelValues(fmt.Sprint(n)).Inc()
		requestLatency.WithLabelValues(fmt.Sprint(n)).Observe(float64(time.Since(start)) / float64(time.Millisecond))
	}
	acct := c.Param("acct")

	var entry atdb.AccountInfo
	if strings.HasPrefix(acct, "did:") {
		// it looks like did
		err := s.r.DB.Model(&entry).Where("did = ?", acct).Limit(1).Take(&entry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			updateMetrics(http.StatusNotFound)
			return c.String(http.StatusBadRequest, "Unknown DID")
		}
	} else {
		// otherwise assume a handle
		err := s.r.DB.Model(&entry).Where("handle = ?", acct).Limit(1).Take(&entry).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			updateMetrics(http.StatusNotFound)
			return c.String(http.StatusBadRequest, "Unknown Handle")
		}
	}

	view := atdb.AccountViewFromInfo(&entry)

	updateMetrics(http.StatusOK)
	return c.JSON(http.StatusOK, view)
}

func (s *Server) Autocomplete(c echo.Context) error {
	start := time.Now()
	updateMetrics := func(c int) {
		requestCount.WithLabelValues(fmt.Sprint(c)).Inc()
		requestLatency.WithLabelValues(fmt.Sprint(c)).Observe(float64(time.Since(start)) / float64(time.Millisecond))
	}
	token := c.Param("token")

	var entries []atdb.AccountInfo
	// assume a handle (todo, profile names, preference following / graph)
	res := s.r.DB.Model(&atdb.AccountInfo{}).Where("handle LIKE ?", token+"%").Limit(10).Find(&entries)
	fmt.Println(res)
	err := res.Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		updateMetrics(http.StatusNotFound)
		return c.String(http.StatusBadRequest, "Unknown Handle")
	}

	var views []atdb.AccountInfoView
	for _, entry := range entries {
		view := atdb.AccountViewFromInfo(&entry)
		views = append(views, view)
	}

	updateMetrics(http.StatusOK)
	return c.JSON(http.StatusOK, views)
}

func (s *Server) DidDoc(c echo.Context) error {

	start := time.Now()
	updateMetrics := func(c int) {
		requestCount.WithLabelValues(fmt.Sprint(c)).Inc()
		requestLatency.WithLabelValues(fmt.Sprint(c)).Observe(float64(time.Since(start)) / float64(time.Millisecond))
	}

	log := zerolog.Ctx(s.r.Ctx)

	requestedDid := c.Param("did")

	// lookup entry in db
	var entry atdb.PLCLogEntry
	err := s.r.DB.Model(&entry).Where("did = ? AND (NOT nullified)", requestedDid).Order("plc_timestamp desc").Limit(1).Take(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		updateMetrics(http.StatusNotFound)
		return c.String(http.StatusNotFound, "unknown DID")
	}
	if err != nil {
		log.Error().Err(err).Str("did", requestedDid).Msgf("Failed to get the last log entry for %q: %s", requestedDid, err)
		updateMetrics(http.StatusInternalServerError)
		return c.String(http.StatusInternalServerError, "failed to get the last log entry")
	}

	// check if account deleted
	if _, ok := entry.Operation.Value.(plc.Tombstone); ok {
		updateMetrics(http.StatusNotFound)
		return c.String(http.StatusNotFound, "DID Deleted")
	}

	// handle legacy
	var op plc.Op
	switch v := entry.Operation.Value.(type) {
	case plc.Op:
		op = v
	case plc.LegacyCreateOp:
		op = v.AsUnsignedOp()
	}

	doc, err := plc.MakeDoc(db.PLCLogEntryToOp(entry), op)
	if err != nil {
		log.Error().Err(err).Str("did", requestedDid).Msgf("Failed to create DID document")
		updateMetrics(http.StatusInternalServerError)
		return c.String(http.StatusInternalServerError, "failed to create DID document")
	}

	updateMetrics(http.StatusOK)
	return c.JSON(http.StatusOK, doc)
}

func mapSlice[A any, B any](s []A, fn func(A) B) []B {
	r := make([]B, 0, len(s))
	for _, v := range s {
		r = append(r, fn(v))
	}
	return r
}
