package db

import (
	"strings"

	"github.com/blebbit/atmunge/pkg/plc"
)

func PLCLogEntryFromOp(op plc.OperationLogEntry) PLCLogEntry {
	return PLCLogEntry{
		DID:          op.DID,
		CID:          op.CID,
		PLCTimestamp: op.CreatedAt,
		Nullified:    op.Nullified,
		Operation:    op.Operation,
	}
}

func PLCLogEntryToOp(entry PLCLogEntry) plc.OperationLogEntry {
	return plc.OperationLogEntry{
		DID:       entry.DID,
		CID:       entry.CID,
		CreatedAt: entry.PLCTimestamp,
		Nullified: entry.Nullified,
		Operation: entry.Operation,
	}
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
