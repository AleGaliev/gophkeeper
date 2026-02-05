package handlers

import (
	"context"
	"encoding/json"
	"gophkeeper/internal/middleware"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

type service interface {
	CreateLoginPass(ctx context.Context, user string, message []byte) error
	СreateSecretText(ctx context.Context, user string, message []byte) error
	CreateBinaryData(ctx context.Context, user string, message []byte) error
	CreateBankCard(ctx context.Context, user string, message []byte) error
	CreateUser(message []byte) (string, error)
	AuthUser(message []byte) (string, error)
	GetLoginFromToken(tokenString string) (string, error)
	GetSecretList(user string) ([]byte, error)
	GetSecret(ctx context.Context, user, secretType, secretName string) ([]byte, error)
}

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
}

type Handler struct {
	service service
	logger  Logger
}

func New(service service, logger Logger) http.Handler {
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
	muxWithMiddlewares.With(middleware.AuthMiddleware(h.service)).Route("/v1/secret/create", func(r chi.Router) {
		r.Post("/userpassword", h.createLoginPasswordSecret)
		r.Post("/text", h.createTextSecret)
		r.Post("/bankcard", h.createBankCardSecret)
		r.Post("/file", h.createFileSecret)
	})
	muxWithMiddlewares.With(middleware.AuthMiddleware(h.service)).Route("/v1/secret/get", func(r chi.Router) {
		r.Get("/list", h.getSecretList)
		r.Get("/{type}/{name}", h.getSecret)
	})

	return muxWithMiddlewares
}

func (h *Handler) userRegister(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	token, err := h.service.CreateUser(requestData.bodyBytes)
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
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	token, err := h.service.AuthUser(requestData.bodyBytes)
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

func (h *Handler) createLoginPasswordSecret(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.service.CreateLoginPass(ctx, requestData.username, requestData.bodyBytes); err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	successResponse(w)
}

func (h *Handler) createTextSecret(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.service.СreateSecretText(ctx, requestData.username, requestData.bodyBytes); err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	successResponse(w)
}

func (h *Handler) createBankCardSecret(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.service.CreateBankCard(ctx, requestData.username, requestData.bodyBytes); err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	successResponse(w)
}

func (h *Handler) createFileSecret(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := h.service.CreateBinaryData(ctx, requestData.username, requestData.bodyBytes); err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	successResponse(w)
}

func (h *Handler) getSecretList(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := h.service.GetSecretList(requestData.username)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *Handler) getSecret(w http.ResponseWriter, r *http.Request) {
	requestData, err := handleRequestData(r)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	typeSecret := chi.URLParam(r, "type")
	nameSecret := chi.URLParam(r, "name")
	jsonBytes, err := h.service.GetSecret(r.Context(), requestData.username, typeSecret, nameSecret)
	if err != nil {
		failedResponse(w, err.Error(), "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json ")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
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
