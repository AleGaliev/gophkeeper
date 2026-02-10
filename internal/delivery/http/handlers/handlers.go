package handlers

import (
	"context"
	"encoding/json"
	"gophkeeper/internal/dto/log"
	"gophkeeper/internal/dto/model"
	"gophkeeper/internal/dto/service"
	"gophkeeper/internal/middleware"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service service.Service
	logger  log.Logger
}

func New(service service.Service, logger log.Logger) http.Handler {
	h := &Handler{
		service: service,
		logger:  logger,
	}
	mux := chi.NewRouter()
	muxWithMiddlewares := mux.With(
		middleware.MiddlewareHandlerLogger(logger),
	)
	muxWithMiddlewares.Route("/", func(r chi.Router) {
		r.Mount("/debug/pprof", pprofRoutes())
	})
	muxWithMiddlewares.Route("/v1/user", func(r chi.Router) {
		r.Post("/register", h.userRegister)
		r.Post("/auth", h.userAuth)
	})
	muxWithMiddlewares.With(middleware.AuthMiddleware(h.service)).Route("/v1/secret/get", func(r chi.Router) {
		r.Get("/list", h.getSecretList)
		r.Get("/secret_type", h.getSecretListInType)
	})

	return muxWithMiddlewares
}

func (h *Handler) userRegister(w http.ResponseWriter, r *http.Request) {

	data := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var user model.User
	if err := data.Decode(&user); err != nil {
		failedResponse(w, err.Error(), "", http.StatusBadRequest)
		return
	}
	token, err := h.service.CreateUser(context.Background(), user)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusForbidden)
		return
	}
	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		MaxAge:   3600,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	successResponse(w)
}

func (h *Handler) userAuth(w http.ResponseWriter, r *http.Request) {
	data := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var user model.User
	if err := data.Decode(&user); err != nil {
		failedResponse(w, err.Error(), "", http.StatusBadRequest)
		return
	}

	token, err := h.service.AuthUser(context.Background(), user)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusForbidden)
		return
	}
	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		MaxAge:   3600,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	successResponse(w)
}

func (h *Handler) getSecretList(w http.ResponseWriter, r *http.Request) {
	user, err := handleGetUser(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	secrets, err := h.service.GetSecretList(context.Background(), user, "")
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(secrets)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (h *Handler) getSecretListInType(w http.ResponseWriter, r *http.Request) {
	user, err := handleGetUser(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusUnauthorized)
		return
	}

	secrets, err := h.service.GetSecretList(context.Background(), user, "")
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(secrets)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func pprofRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", pprof.Index)
	r.Get("/cmdline", pprof.Cmdline)
	r.Get("/profile", pprof.Profile)
	r.Get("/symbol", pprof.Symbol)
	r.Get("/trace", pprof.Trace)
	r.Handle("/goroutine", pprof.Handler("goroutine"))
	r.Handle("/heap", pprof.Handler("heap"))
	r.Handle("/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/block", pprof.Handler("block"))
	r.Handle("/mutex", pprof.Handler("mutex"))
	r.Handle("/allocs", pprof.Handler("allocs"))
	return r
}

func failedResponse(res http.ResponseWriter, message, hashKey string, statusCode int) {
	res.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  `failed`,
		"message": message,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		failedResponse(res, err.Error(), hashKey, statusCode)

	}

	res.WriteHeader(statusCode)
	_, err = res.Write(jsonBytes)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func successResponse(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  `success`,
		"message": "Запрос обработан",
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
	_, err = res.Write(jsonBytes)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	}
}
