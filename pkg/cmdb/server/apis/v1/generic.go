package v1

import (
	"fmt"
	"gcmdb/global"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/conversion"
	"gcmdb/pkg/cmdb/server/storage"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const PathPrefix = "/api/v1"

var db *storage.Store

func InstallApi(r *chi.Mux, s *storage.Store) {
	if s == nil {
		db = newStorage()
	} else {
		db = s
	}

	r.Get(path.Join(PathPrefix, "health"), healthFunc())

	for _, kind := range global.ResourceOrder {
		addGenericApi(r, kind)
	}

	addAppRenderApi(r)
}

func addGenericApi(r *chi.Mux, kind string) {
	kind = strings.ToLower(kind)
	obj, err := cmdb.NewResourceWithKind(kind)
	if err != nil {
		panic(err)
	}
	namespaced := obj.GetMeta().HasNamespace()
	basePath := fmt.Sprintf("%s/%ss", PathPrefix, kind)
	namespacedPath := fmt.Sprintf("%s/%ss/%s", PathPrefix, kind, "{namespace}")

	if namespaced {
		r.Route(basePath, func(r chi.Router) {
			r.Post("/", createFunc(kind))
			r.Get("/", getListFunc(kind))
		})
	}

	r.Get(basePath+"/count/", countFunc(kind))
	r.Get(basePath+"/names/", getNamesFunc(kind))

	if namespaced {
		basePath = namespacedPath
	}
	r.Route(basePath, func(r chi.Router) {
		r.Post("/", createFunc(kind))
		r.Get("/", getListFunc(kind))
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", getFunc(kind))
			r.Post("/", updateFunc(kind))
			r.Delete("/", deleteFunc(kind))
		})
	})
}

func healthFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db.Health(r.Context()) {
			render.Status(r, http.StatusOK)
			render.Respond(w, r, "health")
		}
		render.Status(r, http.StatusFailedDependency)
		render.Respond(w, r, "unhealth")
	}
}

func getFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		revision := r.URL.Query().Get("revision")
		opts := storage.GetOptions{ResourceVersion: revision}

		var out cmdb.Object
		if err := db.Get(r.Context(), kind, name, namespace, opts, &out); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, out)
	}
}

func countFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := r.URL.Query().Get("namespace")

		cnt, err := db.Count(r.Context(), kind, namespace)
		if err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, cnt)
	}
}

func getNamesFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := r.URL.Query().Get("namespace")

		names, err := db.GetNames(r.Context(), kind, namespace)
		if err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, names)
	}
}

func getListFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		var page, limit int
		labelSelector := conversion.ParseSelector(r.URL.Query().Get("selector"))
		fieldSelector := conversion.ParseSelector(r.URL.Query().Get("filed_selector"))
		page, _ = strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		all, _ := strconv.ParseBool(r.URL.Query().Get("all"))
		opts := storage.ListOptions{
			Page: int64(page), Limit: int64(limit), All: all,
			LabelSelector: labelSelector, FieldSelector: fieldSelector,
		}

		var out = []cmdb.Object{}
		if err := db.GetList(r.Context(), kind, namespace, opts, &out); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, out)
	}
}

func createFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := cmdb.NewResourceWithKind(kind)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		if err := render.Decode(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var out cmdb.Object
		if err := db.Create(r.Context(), data, &out); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusCreated)
		render.Respond(w, r, out)
	}
}

func updateFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := cmdb.NewResourceWithKind(kind)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		if err := render.Decode(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var out cmdb.Object
		if err := db.Update(r.Context(), data, &out); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		if out == nil {
			render.Status(r, http.StatusNoContent)
			render.Respond(w, r, nil)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, out)
	}
}

func deleteFunc(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")

		if err := db.Delete(r.Context(), kind, name, namespace); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusNoContent)
		render.Respond(w, r, nil)
	}
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrNotFound(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 404,
		StatusText:     "Not found.",
		ErrorText:      err.Error(),
	}
}

func ErrUnprocessableEntity(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Unprocessable Entity.",
		ErrorText:      err.Error(),
	}
}

func ErrInternal(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal Server error.",
		ErrorText:      err.Error(),
	}
}

func handleStorageErr(w http.ResponseWriter, r *http.Request, err error) {
	switch err := err.(type) {
	case *storage.StorageError:
		switch err.Code {
		case storage.ErrCodeKeyNotFound:
			render.Render(w, r, ErrNotFound(err))
		case storage.ErrCodeInvalidObj:
			render.Render(w, r, ErrUnprocessableEntity(err))
		default:
			render.Render(w, r, ErrInvalidRequest(err))
		}
	default:
		render.Render(w, r, ErrInternal(err))
	}
}

func newStorage() *storage.Store {
	endpoint := global.ServerSetting.ETCD_SERVER_HOST + ":" + global.ServerSetting.ETCD_SERVER_PORT
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
	})
	if err != nil {
		panic(err)
	}
	store := storage.New(client, global.StoragePathPrefix)
	return store
}
