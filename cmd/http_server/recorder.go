package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

//FileRecorder serves as mini db
type FileRecorder interface {
	Exist(string) bool
	AddRecord(string, int64)
	GetRecords() Records
	Close()
}

//Records key is filename, value is size in bytes
type Records map[string]int64

type fileRecorder struct {
	mutex       sync.Mutex
	records     Records
	recordsFile *os.File
}

//Close closes the recordsFile
func (r *fileRecorder) Close() {
	r.recordsFile.Close()
}

//Exist check if file exists in Records
func (r *fileRecorder) Exist(filename string) bool {
	_, x := r.records[filename]
	return x
}

//AddRecord write new record to the recordsFile and to records map
func (r *fileRecorder) AddRecord(filename string, size int64) {
	const op = "AddRecord"

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.records[filename] = size

	log.Info().Msgf("new record %q size=%d", filename, size)
	w := csv.NewWriter(r.recordsFile)
	err := w.Write([]string{filename, strconv.FormatInt(size, 10)})
	w.Flush()
	if err != nil {
		log.Error().Err(err).Msg(op)
	}

}

func (r *fileRecorder) RemoveRecord(filename string, size int64) {
	panic("unimplemented :(")
}

//GetRecords returns all records
func (r *fileRecorder) GetRecords() Records {
	return r.records
}

//NewRecorder instances new recorder
func NewRecorder(csvFilename string) FileRecorder {
	const op = "NewRecorder"

	r := &fileRecorder{records: Records{}}

	err := r.initRecords(csvFilename)
	if err != nil {
		log.Error().Err(err).Msg(op)
	}
	return r
}

//init records will read all existing file records from csvFilename
//if such file not exists, it will create it
func (r *fileRecorder) initRecords(csvFilename string) error {
	const op = "initRecords"

	f, err := os.OpenFile(csvFilename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Error().Err(err).Msg(op)
		return err
	}

	rd := csv.NewReader(f)

	recs, err := rd.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg(op)
	}

	for _, rec := range recs {
		log.Info().Msgf("read record %s size=%s", rec[0], rec[1])
		size, err := strconv.ParseInt(rec[1], 10, 64)
		if err != nil {
			log.Error().Err(err).Msg(op)
		}
		r.records[rec[0]] = size
	}

	r.recordsFile = f

	return err
}
