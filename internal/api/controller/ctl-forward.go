package controller

import (
	"github.com/helays/ssh-proxy-plus/internal/api/service"
	"github.com/helays/ssh-proxy-plus/internal/cache"
	"github.com/helays/ssh-proxy-plus/internal/model"
	"net/http"

	"github.com/helays/utils/v2/net/http/response"
	"golang.org/x/net/websocket"
)

func (c *Controller) CtlForward(w http.ResponseWriter, r *http.Request) {
	serv := service.NewForWardService()
	switch r.Method {
	case http.MethodGet:
		cache.ConnectList.ReadWith(func(connects []model.Connect) {
			response.SetReturnData(w, 0, connects)
		})
	case http.MethodPost:
		if err := serv.Create(r); err != nil {
			response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		if err := serv.Update(r); err != nil {
			response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		if err := serv.Delete(r.URL.Query().Get("id")); err != nil {
			response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
			return
		}
	case http.MethodPatch:
		if err := serv.Stop(r); err != nil {
			response.SetReturnErrorDisableLog(w, err, http.StatusInternalServerError)
			return
		}
	default:
		response.MethodNotAllow(w)
		return
	}
	response.SetReturnData(w, 0, "success")
}

func (c *Controller) CtlWSDataApi(ws *websocket.Conn) {
	serv := service.NewWS(ws)
	serv.Service(ws.Request().Context())
}
