package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/core-go/core"
	s "github.com/core-go/search"

	"go-service/internal/user/model"
	"go-service/internal/user/service"
)

func NewUserHandler(service service.UserService, validate func(context.Context, interface{}) ([]core.ErrorMessage, error), logError func(context.Context, string, ...map[string]interface{}), action *core.ActionConf) *UserHandler {
	userType := reflect.TypeOf(model.User{})
	paramIndex, filterIndex, csvIndex, _ := s.CreateParams(reflect.TypeOf(model.UserFilter{}), userType)
	params := core.CreateParameters(userType, logError, validate, action, paramIndex, filterIndex, csvIndex)
	return &UserHandler{service: service, Parameters: params}
}

type UserHandler struct {
	service service.UserService
	*core.Parameters
}

func (h *UserHandler) All(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.All(r.Context())
	if err != nil {
		h.Error(r.Context(), fmt.Sprintf("Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	core.JSON(w, http.StatusOK, users)
}
func (h *UserHandler) Load(w http.ResponseWriter, r *http.Request) {
	id := core.GetRequiredParam(w, r)
	if len(id) > 0 {
		user, err := h.service.Load(r.Context(), id)
		if err != nil {
			h.Error(r.Context(), fmt.Sprintf("Error to get user %s: %s", id, err.Error()))
			http.Error(w, core.InternalServerError, http.StatusInternalServerError)
			return
		}
		core.JSON(w, core.IsFound(user), user)
	}
}
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user model.User
	er1 := core.Decode(w, r, &user)
	if er1 == nil {
		errors, er2 := h.Validate(r.Context(), &user)
		if !core.HasError(w, r, errors, er2, h.Error, h.Log, h.Resource, h.Action.Create) {
			res, er3 := h.service.Create(r.Context(), &user)
			core.AfterCreated(w, r, &user, res, er3, h.Error, h.Log, h.Resource, h.Action.Create)
		}
	}
}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var user model.User
	er1 := core.DecodeAndCheckId(w, r, &user, h.Keys, h.Indexes)
	if er1 == nil {
		errors, er2 := h.Validate(r.Context(), &user)
		if !core.HasError(w, r, errors, er2, h.Error, h.Log, h.Resource, h.Action.Update) {
			res, er3 := h.service.Update(r.Context(), &user)
			core.HandleResult(w, r, &user, res, er3, h.Error, h.Log, h.Resource, h.Action.Update)
		}
	}
}
func (h *UserHandler) Patch(w http.ResponseWriter, r *http.Request) {
	var user model.User
	r, jsonUser, er1 := core.BuildMapAndCheckId(w, r, &user, h.Keys, h.Indexes)
	if er1 == nil {
		errors, er2 := h.Validate(r.Context(), &user)
		if !core.HasError(w, r, errors, er2, h.Error, h.Log, h.Resource, h.Action.Patch) {
			res, er3 := h.service.Patch(r.Context(), jsonUser)
			core.HandleResult(w, r, jsonUser, res, er3, h.Error, h.Log, h.Resource, h.Action.Patch)
		}
	}
}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := core.GetRequiredParam(w, r)
	if len(id) > 0 {
		res, err := h.service.Delete(r.Context(), id)
		core.HandleDelete(w, r, res, err, h.Error, h.Log, h.Resource, h.Action.Delete)
	}
}
func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	filter := model.UserFilter{Filter: &s.Filter{}}
	s.Decode(r, &filter, h.ParamIndex, h.FilterIndex)

	offset := s.GetOffset(filter.Limit, filter.Page)
	users, total, err := h.service.Search(r.Context(), &filter, filter.Limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	core.JSON(w, http.StatusOK, &s.Result{List: &users, Total: total})
}
