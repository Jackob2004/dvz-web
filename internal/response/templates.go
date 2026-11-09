package response

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/Jackob2004/dvz-web/assets"
	"github.com/Jackob2004/dvz-web/internal/funcs"
)

func Page(w http.ResponseWriter, status int, data any, pagePath string) error {
	return PageWithHeaders(w, status, data, nil, pagePath)
}

func PageWithHeaders(w http.ResponseWriter, status int, data any, headers http.Header, pagePath string) error {
	patterns := []string{"base.gohtml", "partials/*.gohtml", pagePath}

	return NamedTemplateWithHeaders(w, status, data, headers, "base", patterns...)
}

func NamedTemplate(w http.ResponseWriter, status int, data any, templateName string, patterns ...string) error {
	return NamedTemplateWithHeaders(w, status, data, nil, templateName, patterns...)
}

func NamedTemplateWithHeaders(w http.ResponseWriter, status int, data any, headers http.Header, templateName string, patterns ...string) error {
	for i := range patterns {
		patterns[i] = "templates/" + patterns[i]
	}

	ts, err := template.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, patterns...)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	err = ts.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		return err
	}

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(status)
	buf.WriteTo(w)

	return nil
}

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

func NamedSSETemplate(w http.ResponseWriter, buffer *bytes.Buffer, data any, templateName, pattern string) error {
	buffer.Reset()
	ts, err := template.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, "templates/"+pattern)
	if err != nil {
		return err
	}

	sseWriter := newSSEWriter(buffer)

	err = ts.ExecuteTemplate(sseWriter, templateName, data)
	if err != nil {
		return err
	}

	return sseWriter.FlushEvent(w)
}
