package v1

import (
	"gcmdb/pkg/cmdb/server/storage"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func NewRouter(s *storage.Store) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Timeout(3 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	InstallApi(r, s)

	return r
}
