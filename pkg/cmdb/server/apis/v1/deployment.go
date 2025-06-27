package v1

import (
	"fmt"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/deployment"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type RenderParams struct {
	Params map[string]any `json:"params"`
}

// TODO: read appdeployment status
// TODO: read logs
// TODO: list appdeployment image tags

func addAppRenderApi(r *chi.Mux) {
	r.Post(
		fmt.Sprintf("%s/appdeployments/{namespace}/{name}/render", PathPrefix),
		renderAppDeploymentFunc(),
	)
	r.Post(
		fmt.Sprintf("%s/appdeployments/{namespace}/{name}/deploytemplate/render", PathPrefix),
		renderDeployTemplateFunc(),
	)
	r.Post(
		fmt.Sprintf("%s/appdeployments/{namespace}/{name}/run/{action}", PathPrefix),
		runAppDeploymentFunc(),
	)
}

// render appdeployment
func renderAppDeploymentFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		var params RenderParams
		if err = render.Decode(r, &params); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var appDeploy *cmdb.AppDeployment
		if appDeploy, err = deployment.ResolveAppDeployment(db, name, namespace, params.Params); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, appDeploy)
	}
}

// render deploytemplate
func renderDeployTemplateFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		var params RenderParams
		if err = render.Decode(r, &params); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		var deployTemplate *cmdb.DeployTemplate
		if deployTemplate, err = deployment.ResolveDeployTemplate(db, name, namespace, params.Params); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, deployTemplate)
	}
}

// TODO: run appdeployment
func runAppDeploymentFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		name := chi.URLParam(r, "name")
		namespace := chi.URLParam(r, "namespace")
		action := chi.URLParam(r, "action")
		var params RenderParams
		if err = render.Decode(r, &params); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		deployCtl := deployment.NewDeployController(
			db,
			deployment.DeployAction(action),
			name,
			namespace,
			params.Params,
		)
		var appDeploy *cmdb.AppDeployment
		if appDeploy, err = deployCtl.Run(); err != nil {
			handleStorageErr(w, r, err)
			return
		}
		render.Status(r, http.StatusOK)
		render.Respond(w, r, appDeploy)
	}
}
