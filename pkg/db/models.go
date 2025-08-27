package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/blebbit/atmunge/pkg/plc"
)

// should be same as type of gorm.Model.ID
type ID uint

// JSONRaw defined to handle raw JSON data
type JSONRaw json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSONRaw) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*j = bytes
	return nil
}

// Value return json value, implement driver.Valuer interface
func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// this came from original plc-mirror code base
type PLCLogEntryUnique struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time

	DID          string        `gorm:"column:did;index:did_timestamp;uniqueIndex:did_cid"`
	CID          string        `gorm:"column:cid;uniqueIndex:did_cid"`
	PLCTimestamp string        `gorm:"column:plc_timestamp;index:did_timestamp,sort:desc;index:,sort:desc"`
	Nullified    bool          `gorm:"default:false"`
	Operation    plc.Operation `gorm:"type:JSONB;serializer:json"`
}

type PLCLogEntry struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time

	DID          string        `gorm:"column:did;index"`
	CID          string        `gorm:"column:cid;index"`
	PLCTimestamp string        `gorm:"column:plc_timestamp;index:,sort:desc"`
	Nullified    bool          `gorm:"default:false"`
	Operation    plc.Operation `gorm:"type:JSONB;serializer:json"`

	// custom notes on the log entry, mainly for describing issues and errors
	Notes    string `gorm:"column:notes"`
	Filtered int    `gorm:"column:filtered;default:0"` // roughly the number of issues
}

type PdsRepo struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	PDS    string `gorm:"column:pds;index:idx_pds"`
	DID    string `gorm:"column:did;index:idx_did"`
	Head   string `gorm:"column:head"`
	Rev    string `gorm:"column:rev"`
	Active bool   `gorm:"column:active;default:true"`
	Status string `gorm:"column:status"`
}

type AccountInfo struct {
	ID        ID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	DID          string `gorm:"column:did;index:did_timestamp;uniqueIndex:did"`
	PLCTimestamp string `gorm:"column:plc_timestamp;index:did_timestamp,sort:desc;index:,sort:desc"`

	PDS    string `gorm:"column:pds"`
	Handle string `gorm:"column:handle;index:idx_handle"`

	HandleMatch bool `gorm:"handle_match"`
	// when did we last check if the handle points at the DID?
	HandleMatchLastChecked time.Time

	// extra info
	Describe JSONRaw `gorm:"column:describe;type:JSONB"`
	Extra    JSONRaw `gorm:"column:extra;type:JSONB"`
}

type AccountRepo struct {
	DID string `gorm:"primarykey;column:did;index:did_timestamp;uniqueIndex:did"`

	Rev         string `gorm:"column:rev;index:idx_rev"`
	LastChanged time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

const PlcLogEntryConflictPsqlfunction = `
create or replace function before_update_on_plc_log_entries()
returns trigger language plpgsql as $$
begin
    if new.did = old.did AND new.cid = old.cid then
        insert into plc_log_entry_conflicts(oid, did, cid, updated_at)
        values (old.id, old.did, old.cid, now())
        on conflict(old.id, did, cid)
        do update set updated_at = now(), count = plc_log_entry_conflicts.count + 1;
        return null;
    end if;
    return new;
end $$;

create trigger before_update_on_plc_log_entries
before update on plc_log_entries
for each row execute procedure before_update_on_plc_log_entries();
`
const PLCLogEntryConflictPsqlException = `
create or replace function insert_or_log(arg_content text)
returns void language plpgsql as
$$
begin
  insert into fiche (content) values (arg_content);
exception 
  when unique_violation then
    insert into fiche_qualites (poi_id, qualites_id)
    values ((select idPoi from fiche where content = arg_content), 'Code error');
end;
$$;
`
