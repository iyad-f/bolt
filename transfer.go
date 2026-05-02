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
	"bufio"
	"io"
	"strconv"
	"strings"
)

type chunkReader struct {
	r         *bufio.Reader
	remaining int64
	done      bool
}

func (cr *chunkReader) Read(p []byte) (int, error) {
	if cr.done {
		return 0, io.EOF
	}

	if cr.remaining == 0 {
		line, err := cr.r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimRight(line, "\r\n")

		size, err := strconv.ParseInt(line, 16, 64)
		if err != nil {
			return 0, err
		}

		if size == 0 {
			cr.done = true
			return 0, io.EOF
		}

		cr.remaining = size
	}

	toRead := min(int64(len(p)), cr.remaining)

	n, err := cr.r.Read(p[:toRead])
	cr.remaining -= int64(n)

	if cr.remaining == 0 {
		cr.r.ReadString('\n')
	}

	return n, err
}

type chunkWriter struct {
	w io.Writer
}

func (cw *chunkWriter) Write(p []byte) (int, error) {
	_, err := cw.w.Write([]byte(strconv.FormatInt(int64(len(p)), 16)))
	if err != nil {
		return 0, err
	}
	_, err = cw.w.Write(crlf)
	if err != nil {
		return 0, err
	}
	n, err := cw.w.Write(p)
	if err != nil {
		return n, err
	}
	_, err = cw.w.Write(crlf)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (cw *chunkWriter) Close() error {
	_, err := cw.w.Write([]byte("0\r\n\r\n"))
	return err
}
