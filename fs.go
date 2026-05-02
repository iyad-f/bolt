// Copyright 2026 Iyad
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bolt

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func fileServer(root string) HandlerFunc {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic("invalid root path: " + err.Error())
	}

	return func(w ResponseWriter, r *Request) {
		cleanPath := filepath.Clean("/" + r.PathValue("filepath"))
		fullPath := filepath.Join(absRoot, cleanPath)

		// Block access outside the root directory (prevent path traversal).
		if !strings.HasPrefix(fullPath, absRoot) {
			w.WriteHeader(StatusForbidden)
			w.Write([]byte(StatusText(StatusForbidden)))
			return
		}

		file, fileInfo, err := openFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(StatusNotFound)
				w.Write([]byte(StatusText(StatusNotFound)))
			} else {
				w.WriteHeader(StatusInternalServerError)
				w.Write([]byte(StatusText(StatusInternalServerError)))
			}
			return
		}
		defer file.Close()

		if !isFileModified(r, fileInfo) {
			w.WriteHeader(StatusNotModified)
			return
		}

		w.Header().Set("ETag", generateETag(fileInfo))
		w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(time.RFC1123))
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Type", detectContentType(fileInfo.Name()))

		if r.Header.Get("Range") != "" {
			serveRange(w, r, file, fileInfo.Size())
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(int(fileInfo.Size())))

		if r.Method == "HEAD" {
			return
		}
		io.Copy(w, file)
	}
}

func openFile(path string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	// if its a directory, try to open the index.html file
	if fileInfo.IsDir() {
		indexFile, err := os.Open(filepath.Join(path, "index.html"))
		if err != nil {
			return nil, nil, err
		}
		file.Close()
		file = indexFile

		fileInfo, err = file.Stat()
		if err != nil {
			return nil, nil, err
		}
	}

	return file, fileInfo, nil
}

func isFileModified(r *Request, fileInfo os.FileInfo) bool {
	etag := generateETag(fileInfo)

	if match := r.Header.Get("If-None-Match"); match != "" {
		return match != etag
	}

	if ims := r.Header.Get("If-Modified-Since"); ims != "" {
		imsTime, err := time.Parse(time.RFC1123, ims)
		if err != nil {
			return true
		}
		return fileInfo.ModTime().After(imsTime)
	}

	return true
}

func generateETag(fileInfo os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, fileInfo.ModTime().Unix(), fileInfo.Size())
}

func detectContentType(name string) string {
	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return contentType
}

func serveRange(w ResponseWriter, r *Request, file *os.File, fileSize int64) {
	start, end, valid := parseRange(r.Header.Get("Range"), fileSize)
	if !valid {
		w.WriteHeader(StatusRequestedRangeNotSatisfiable)
		w.Write([]byte(StatusText(StatusRequestedRangeNotSatisfiable)))
		return
	}

	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.Itoa(int(end-start+1)))

	w.WriteHeader(StatusPartialContent)

	if r.Method != "HEAD" {
		file.Seek(start, io.SeekStart)
		io.CopyN(w, file, end-start+1)
	}
}

func parseRange(header string, fileSize int64) (start int64, end int64, valid bool) {
	remaining, found := strings.CutPrefix(header, "bytes=")
	if !found {
		return 0, 0, false
	}

	parts := strings.Split(remaining, "-")
	if len(parts) != 2 {
		return 0, 0, false
	}

	if parts[0] == "" {
		// suffix range: -b
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, false
		}
		start = fileSize - n
		end = fileSize - 1
	} else {
		var err error

		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, false
		}

		if parts[1] == "" {
			// open end: a-
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, 0, false
			}
		}
	}

	if start < 0 || start >= fileSize || end < start || end >= fileSize {
		return 0, 0, false
	}

	return start, end, true
}
