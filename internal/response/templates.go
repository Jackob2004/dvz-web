package response

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type SSEWriter struct {
	intermediateTarget *bytes.Buffer
	atLineStart        bool
}

func newSSEWriter(intermediateTarget *bytes.Buffer) *SSEWriter {
	return &SSEWriter{
		intermediateTarget: intermediateTarget,
		atLineStart:        true,
	}
}

var ssePrefix = []byte("data: ")
var sseCloseAtStart = []byte("\n")
var sseClose = []byte("\n\n")

var ErrStreamBroken = errors.New("stream broken")

func (w *SSEWriter) Write(p []byte) (int, error) {
	var cursor int

	for cursor < len(p) {
		if w.atLineStart {
			w.intermediateTarget.Write(ssePrefix)
			w.atLineStart = false
		}

		offset := bytes.IndexByte(p[cursor:], '\n')

		if offset == -1 {
			w.intermediateTarget.Write(p[cursor:])
			break
		}

		end := cursor + offset + 1
		w.intermediateTarget.Write(p[cursor:end])
		cursor = end
		w.atLineStart = true
	}

	return len(p), nil
}

func (w *SSEWriter) FlushEvent(finalTarget io.Writer) error {
	term := sseClose
	if w.atLineStart {
		term = sseCloseAtStart
	}

	w.intermediateTarget.Write(term)

	if _, err := w.intermediateTarget.WriteTo(finalTarget); err != nil {
		return fmt.Errorf("%w: %v", ErrStreamBroken, err)
	}

	finalTarget.(http.Flusher).Flush()

	return nil
}

func NamedSSETemplate(w http.ResponseWriter, buffer *bytes.Buffer, data any, t *template.Template, templateName string) error {
	buffer.Reset()

	sseWriter := newSSEWriter(buffer)

	err := t.ExecuteTemplate(sseWriter, templateName, data)
	if err != nil {
		return err
	}

	return sseWriter.FlushEvent(w)
}
