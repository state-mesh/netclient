package functions

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/devilcove/httpclient"
	"github.com/gravitl/netclient/auth"
	"github.com/gravitl/netclient/config"
	"github.com/gravitl/netclient/daemon"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
)

// Pull - pulls the latest config from the server, if manual it will overwrite
func Pull() error {

	serverName := config.CurrServer
	server := config.GetServer(serverName)
	if server == nil {
		return errors.New("server config not found")
	}
	token, err := auth.Authenticate(server, config.Netclient())
	if err != nil {
		return err
	}
	endpoint := httpclient.JSONEndpoint[models.HostPull, models.ErrorResponse]{
		URL:           "https://" + server.API,
		Route:         "/api/v1/host",
		Method:        http.MethodGet,
		Authorization: "Bearer " + token,
		Response:      models.HostPull{},
		ErrorResponse: models.ErrorResponse{},
	}
	pullResponse, errData, err := endpoint.GetJSON(models.HostPull{}, models.ErrorResponse{})
	if err != nil {
		if errors.Is(err, httpclient.ErrStatus) {
			logger.Log(0, "error pulling server", serverName, strconv.Itoa(errData.Code), errData.Message)
		}
		return err
	}
	_ = config.UpdateHostPeers(server.Server, pullResponse.Peers)
	pullResponse.ServerConfig.MQPassword = server.MQPassword // pwd can't change currently
	config.UpdateServerConfig(&pullResponse.ServerConfig)
	fmt.Printf("completed pull for server %s\n", serverName)

	_ = config.WriteServerConfig()
	_ = config.WriteNetclientConfig()
	logger.Log(3, "restarting daemon")
	return daemon.Restart()
}
