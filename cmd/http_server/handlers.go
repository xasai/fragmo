package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
)

func LoggerMware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Info().Str("URI", req.RequestURI).Msg(req.Method)
		next.ServeHTTP(w, req)
	})
}

///////////////////////
// HANDLERS
///////////////////////
func listHandler(w http.ResponseWriter, req *http.Request) {
	const op = "listHandler"

	err := templates.ExecuteTemplate(w, "index.html", recorder.GetRecords())
	if err != nil {
		log.Error().Err(err).Msg(op + ".ExecuteTemplate")
	}
}
func uploadHandler(w http.ResponseWriter, req *http.Request) {
	const op = "uploadHandler"

	if req.Method == "POST" {
		upload(w, req)
	}

	if err := templates.ExecuteTemplate(w, "upload.html", nil); err != nil {
		log.Error().Err(err).Msg(op + ".ExecuteTemplate")
	}
}

func downloadHandler(w http.ResponseWriter, req *http.Request) {
	const op = "downloadHandler"
	http.Redirect(w, req, "/", http.StatusMovedPermanently)
}

///////////////////////
// UPLOAD LOGIC
///////////////////////
//uplaod read all files in separate goroutines and send them to the storage_server by stream
//It reads no more than 1 Mb from file at once and fill last chunk of file with zeroes if its size less than 1 Mb
func upload(w http.ResponseWriter, req *http.Request) {
	const op = "upload"

	req.ParseMultipartForm(32 << 20)
	fileHeaders := req.MultipartForm.File["files"]

	wg := &sync.WaitGroup{}
	errCh := make(chan error, len(fileHeaders))

	for _, h := range fileHeaders {
		h := h

		wg.Add(1)
		go func() {
			defer wg.Done()

			filename := h.Filename
			nameOnly := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
			for idx := 1; ; idx++ {
				if recorder.Exist(filename) {
					//generate new filename if such file already exists
					filename = fmt.Sprintf("%s(%d)%s", nameOnly, idx, filepath.Ext(h.Filename))
					continue
				}
				break
			}

			//Open incoming file
			f, err := h.Open()
			if err != nil {
				errCh <- fmt.Errorf("%v %s", err, filename)
				return
			}
			defer f.Close()

			//gRPC request preparation
			stream, err := storage.Upload(context.Background())
			if err != nil {
				errCh <- fmt.Errorf("%v %s", err, filename)
				return
			}
			reqRPC := &rpc.UploadReq{
				Filename: filename,
				Index:    0,
			}

			size := int64(0)
			buf := make([]byte, cfg.FragmentSize)
			for {

				reqRPC.Index++

				//Read no more than 1 mb
				r, err := f.Read(buf)
				if err != nil && err != io.EOF {
					errCh <- fmt.Errorf("%v %s", err, filename)
					return
				}

				size += int64(r)

				//quit loop on last read fragment
				if size == h.Size {
					break
				}

				reqRPC.Data = buf
				if err = stream.Send(reqRPC); err != nil {
					errCh <- fmt.Errorf("%v %s", err, filename)
					return
				}
			}

			//Fill last fragment with zero bytes t
			left := size % cfg.FragmentSize
			for i := int64(0); i < left; i++ {
				buf[cfg.FragmentSize-left+i] = 0
			}

			reqRPC.Data = buf
			err = stream.Send(reqRPC)
			if err != nil {
				errCh <- fmt.Errorf("%v %s", err, filename)
				return
			}

			_, err = stream.CloseAndRecv()
			if err != nil {
				errCh <- fmt.Errorf("%v %s", err, filename)
				return
			}

			recorder.AddRecord(filename, h.Size)

		}()
	}

	wg.Wait()
	close(errCh)

	wasErr := false

	for err := range errCh {
		wasErr = true
		log.Error().Err(err).Msg("while uploading")
	}
	if wasErr {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

}
