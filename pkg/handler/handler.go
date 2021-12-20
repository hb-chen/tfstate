package handler

import (
	"io"
	"net/http"

	"github.com/hb-chen/tfstate/pkg/state"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Handler interface {
	Get(ctx echo.Context) error
	Update(ctx echo.Context) error
	Lock(ctx echo.Context) error
	Unlock(ctx echo.Context) error
}

type handler struct {
	state state.State
}

func (h *handler) Get(ctx echo.Context) error {
	id := ctx.Param("id")
	d, err := h.state.Get(id)
	if err != nil {
		return err
	}

	return ctx.Blob(http.StatusOK, echo.MIMEApplicationJSON, d)
}

// /state/stackId?ID=e4b05a57-e7fd-26a0-76c1-62b4703d95c0
func (h *handler) Update(ctx echo.Context) error {
	stackId := ctx.Param("stackId")
	log.Infof("update stack: %v", stackId)

	id := ctx.QueryParam("ID")
	log.Infof("update ID: %v", id)

	params, _ := ctx.FormParams()
	log.Infof("body: %++v", params)

	b, _ := io.ReadAll(ctx.Request().Body)
	log.Infof("lock: %v", string(b))
	ctx.Request().Body.Close()

	err := h.state.Update(stackId, b)
	if err != nil {
		return err
	}

	return ctx.JSONBlob(http.StatusOK, []byte{})
}

// {
// 	"Created" : "2021-12-18T08:45:24.730721Z",
// 	"Version" : "1.0.11",
// 	"Operation" : "OperationTypePlan",
// 	"Path" : "",
// 	"Info" : "",
// 	"ID" : "8f9b3687-4100-c0d6-88a4-d6c3cab2abdb",
// 	"Who" : "Steven@StevenChendeMacBook-Pro-4.local"
// }
type LockBody struct {
	Id        string
	Who       string
	Info      string
	Path      string
	Operation string
	Version   string
	Created   string
}

// 423: Locked
// 409: Conflict with the holding lock info when it's already taken
// 200: OK for success

func (h *handler) Lock(ctx echo.Context) error {
	stackId := ctx.Param("stackId")
	log.Infof("local stack: %v", stackId)

	req := &LockBody{}
	if err := ctx.Bind(req); err != nil {
		return nil
	}

	// TODO token validate
	if err := h.state.Lock(stackId, req.Id); err != nil {
		return err
	}

	return nil
}

func (h *handler) Unlock(ctx echo.Context) error {
	stackId := ctx.Param("stackId")
	log.Infof("unlock stack: %v", stackId)

	req := &LockBody{}
	if err := ctx.Bind(req); err != nil {
		return nil
	}

	// TODO token validate
	if err := h.state.Unlock(stackId, req.Id); err != nil {
		return err
	}

	return nil
}

func NewHandler() Handler {
	return &handler{state.NewState()}
}
