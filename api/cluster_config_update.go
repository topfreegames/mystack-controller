package api

import (
	"net/http"

	"github.com/topfreegames/mystack-controller/models"
)

func (c *ClusterConfigHandler) update(w http.ResponseWriter, r *http.Request) {
	logger := loggerFromContext(r.Context())
	clusterName := GetClusterName(r)
	clusterConfig := clusterConfigFromCtx(r.Context())

	log(logger, "Updating config '%s'", clusterName)

	log(logger, "Deleting cluster config '%s'", clusterName)
	err := models.RemoveClusterConfig(c.App.DB, clusterName)
	if err != nil {
		c.App.HandleError(w, Status(err), "removing cluster config error", err)
		return
	}

	log(logger, "Recreating cluster config '%s'", clusterName)
	err = models.WriteClusterConfig(c.App.DB, clusterName, clusterConfig)
	if err != nil {
		c.App.HandleError(w, Status(err), "writing cluster config error", err)
		return
	}

	Write(w, http.StatusOK, `{"status": "ok"}`)
	log(logger, "Config '%s' successfully updated", clusterName)
}
