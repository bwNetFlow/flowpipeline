// Dumps all incoming flow messages to a local sqlite database. The schema used
// for this is preset.
package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Sqlite struct {
	segments.BaseSegment
	db *sql.DB

	FileName string // required
}

// Every Segment must implement a New method, even if there isn't any config
// it is interested in.
func (segment Sqlite) New(config map[string]string) segments.Segment {
	// do config stuff here, add it to fields maybe
	if config["filename"] == "" {
		log.Println("[error] Sqlite: This segment requires a 'filename' parameter.")
		return nil
	}
	_, err := sql.Open("sqlite3", config["filename"])
	if err != nil {
		log.Printf("[error] Sqlite: Could not open DB file at %s.", config["filename"])
		return nil
	}

	return &Sqlite{
		FileName: config["filename"],
	}
}

func (segment *Sqlite) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var err error
	segment.db, err = sql.Open("sqlite3", segment.FileName)
	if err != nil {
		log.Panic(err) // this has already been checked in New
	}
	defer segment.db.Close()

	sqlStmt := `CREATE TABLE IF NOT EXISTS flows (
		Type TEXT not null,
		TimeReceived INTEGER,
		SequenceNum INTEGER,
	        SamplingRate INTEGER,
		SamplerAddress TEXT,
	        TimeFlowStart INTEGER,
	        TimeFlowEnd INTEGER,
	        Bytes INTEGER,
	        Packets INTEGER,
	        SrcAddr TEXT not null,
	        DstAddr TEXT not null,
	        Etype INTEGER,
	        Proto INTEGER not null,
	        SrcPort INTEGER not null,
	        DstPort INTEGER not null,
	        InIf INTEGER not null,
	        OutIf INTEGER,
	        IngressVrfID INTEGER,
	        EgressVrfID INTEGER,
		IPTos INTEGER,
	        ForwardingStatus INTEGER,
	        TCPFlags INTEGER,
	        SrcAS INTEGER,
	        NextHop TEXT,
	        SrcNet INTEGER,
	        DstNet INTEGER,
	        Cid INTEGER,
	        Normalized TEXT,
	        SrcIfName TEXT,
	        SrcIfDesc TEXT,
	        SrcIfSpeed INTEGER,
	        DstIfName TEXT,
	        DstIfDesc TEXT,
	        DstIfSpeed INTEGER,
	        ProtoName TEXT,
	        RemoteCountry TEXT);`
	tx, err := segment.db.Begin()
	if err != nil {
		log.Panicf("[error] Sqlite: Could not start initiation transaction with error: %v", err)
	}
	_, err = tx.Exec(sqlStmt)
	if err != nil {
		log.Panicf("%+v", err) // this should never occur, as it would indicate an error in above create table statement
	}
	tx.Commit()

	for msg := range segment.In {
		sqlStmt := fmt.Sprintf(`INSERT INTO flows VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		  	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

		tx, err := segment.db.Begin()
		if err != nil {
			log.Printf("[warning] Sqlite: Could not start transaction with error: %v", err)
			continue
		}

		stmt, err := tx.Prepare(sqlStmt)
		if err != nil {
			log.Printf("[warning] Sqlite: Could not prepare statement with error: %v", err) // should never happen, indicates an error in above insert statement
			tx.Rollback()
			segment.Out <- msg
			continue
		}
		defer stmt.Close()

		_, err = stmt.Exec(msg.Type, msg.TimeReceived, msg.SequenceNum, msg.SamplingRate,
			fmt.Sprint(net.IP(msg.SamplerAddress)), msg.TimeFlowStart, msg.TimeFlowEnd, msg.Bytes,
			msg.Packets, fmt.Sprint(net.IP(msg.SrcAddr)), fmt.Sprint(net.IP(msg.DstAddr)), msg.Etype, msg.Proto, msg.SrcPort,
			msg.DstPort, msg.InIf, msg.OutIf, msg.IngressVrfID, msg.EgressVrfID,
			msg.IPTos, msg.ForwardingStatus, msg.TCPFlags, msg.SrcAS, fmt.Sprint(net.IP(msg.NextHop)),
			msg.SrcNet, msg.DstNet, msg.Cid, msg.Normalized, msg.SrcIfName,
			msg.SrcIfDesc, msg.SrcIfSpeed, msg.DstIfName, msg.DstIfDesc, msg.DstIfSpeed,
			msg.ProtoName, msg.RemoteCountry)
		if err != nil {
			log.Printf("[warning] Sqlite: Could not insert flow data with error: %v", err)
			tx.Rollback()
			segment.Out <- msg
			continue
		}
		tx.Commit()
		segment.Out <- msg
	}
}

func init() {
	segment := &Sqlite{}
	segments.RegisterSegment("sqlite", segment)
}
