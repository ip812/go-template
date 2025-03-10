package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

// We use Cloudflare tunnels & Traefik proxy in front of the app
// The real user ip address is set to Cf-Connecting-Ip header
// The r.RemoteAddr gets back the ip of Traefik
func RealUserIP(r *http.Request) string {
	return r.Header.Get("Cf-Connecting-Ip")
}

func Render(w http.ResponseWriter, r *http.Request, c templ.Component) error {
	w.Header().Set("Content-Type", "text/html")

	err := c.Render(r.Context(), w)
	if err != nil {
		return ErrorInternalServerError(fmt.Errorf("server failed to render this component"))
	}

	return nil
}

func MakeTemplHandler(ctx context.Context, logger Logger, f func(ctx context.Context, logger Logger, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(ctx, logger, w, r); err != nil {
			var t Toast
			if errors.As(err, &t) {
				AddToast(w, t)
			}
		}
	}
}

func HxRedirect(w http.ResponseWriter, path string) {
	w.Header().Set("HX-Redirect", path)
}
