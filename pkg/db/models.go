package db

import (
	"strings"
	"time"

	"github.com/blebbit/plc-mirror/pkg/plc"
)

type ID uint // should be same as type of gorm.Model.ID

type PLCLogEntry struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time

	DID          string        `gorm:"column:did;index:did_timestamp;uniqueIndex:did_cid"`
	CID          string        `gorm:"column:cid;uniqueIndex:did_cid"`
	PLCTimestamp string        `gorm:"column:plc_timestamp;index:did_timestamp,sort:desc;index:,sort:desc"`
	Nullified    bool          `gorm:"default:false"`
	Operation    plc.Operation `gorm:"type:JSONB;serializer:json"`
}

func PLCLogEntryFromOp(op plc.OperationLogEntry) PLCLogEntry {
	return PLCLogEntry{
		DID:          op.DID,
		CID:          op.CID,
		PLCTimestamp: op.CreatedAt,
		Nullified:    op.Nullified,
		Operation:    op.Operation,
	}
}

type AccountInfo struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	DID          string `gorm:"column:did;index:did_timestamp;uniqueIndex:did"`
	PLCTimestamp string `gorm:"column:plc_timestamp;index:did_timestamp,sort:desc;index:,sort:desc"`

	PDS    string `gorm:"column:pds"`
	Handle string `gorm:"column:handle;uniqueIndex:idx_handle"`

	HandleMatch bool      `gorm:"handle_match"`
	LastChecked time.Time // when did we last check if the handle points at the DID?
}

func AccountInfoFromOp(entry plc.OperationLogEntry) AccountInfo {
	ai := AccountInfo{
		DID:          entry.DID,
		PLCTimestamp: entry.CreatedAt,
	}

	var op plc.Op
	switch v := entry.Operation.Value.(type) {
	case plc.Op:
		op = v
	case plc.LegacyCreateOp:
		op = v.AsUnsignedOp()
	}

	if len(op.AlsoKnownAs) > 0 {
		ai.Handle = strings.TrimPrefix(op.AlsoKnownAs[0], "at://")
	}

	if svc, ok := op.Services["atproto_pds"]; ok {
		ai.PDS = svc.Endpoint
	}
	return ai
}

type AccountInfoView struct {
	DID    string `json:"did"`
	PDS    string `json:"pds"`
	Handle string `json:"handle"`

	PLCTime  string    `json:"plcTime"`
	LastTime time.Time `json:"lastTime"`
}

func AccountViewFromInfo(info *AccountInfo) AccountInfoView {
	return AccountInfoView{
		DID:      info.DID,
		PDS:      info.PDS,
		Handle:   info.Handle,
		PLCTime:  info.PLCTimestamp,
		LastTime: info.UpdatedAt,
	}
}
