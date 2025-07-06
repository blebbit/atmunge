package db

import "time"

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
