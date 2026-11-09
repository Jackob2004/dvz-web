package main

import (
	"fmt"
	"net/http"
)

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Go(func() {
		defer func() {
			pv := recover()
			if pv != nil {
				app.reportServerError(r, fmt.Errorf("%v", pv))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	})
}

func checkETag(r *http.Request, etag string) bool {
	return r.Header.Get("If-None-Match") == etag
}

func deref[T any](ptr *T) T {
	if ptr != nil {
		return *ptr
	}
	var zero T
	return zero
}
