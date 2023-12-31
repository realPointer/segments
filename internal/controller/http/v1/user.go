package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/realPointer/segments/internal/entity"
	"github.com/realPointer/segments/internal/service"
	"github.com/realPointer/segments/pkg/logger"
)

type userRoutes struct {
	userService service.User
}

func NewUserRouter(userService service.User, l logger.Interface) http.Handler {
	u := userRoutes{userService: userService}
	r := chi.NewRouter()

	r.Post("/", u.createUser)
	r.Post("/segments", u.addOrRemoveUserSegments)
	r.Get("/segments", u.getUserSegments)
	r.Get("/operations", u.getUserOperations)
	r.Get("/operations/report-link", u.getUserOperationsYandex)

	return r
}

// @Summary Create user
// @Description Creates a new user with the given ID
// @Tags User
// @Param user_id path int true "user_id"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /user/{user_id} [post]
func (u *userRoutes) createUser(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.userService.CreateUser(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary Get user segments
// @Description Returns a list of segments for the given user
// @Tags User
// @Param user_id path int true "user_id"
// @Success 200 {array} string
// @Failure 400
// @Failure 500
// @Router /user/{user_id}/segments [get]
func (u *userRoutes) getUserSegments(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	segments, err := u.userService.GetUserSegments(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, segments)
}

type Segments struct {
	AddSegments    []entity.AddSegment `json:"add_segments"`
	RemoveSegments []string            `json:"remove_segments"`
}

// @Summary Add or remove user segments
// @Description Adds or removes segments for the given user
// @Tags User
// @Accept json
// @Param user_id path int true "user_id"
// @Param segments body Segments true "segments"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /user/{user_id}/segments [post]
func (u *userRoutes) addOrRemoveUserSegments(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var segments Segments
	err = render.DecodeJSON(r.Body, &segments)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.userService.AddOrRemoveUserSegments(r.Context(), userId, segments.AddSegments, segments.RemoveSegments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get user operations
// @Description Returns a list of operations for the given user
// @Tags User
// @Param user_id path int true "user_id"
// @Param date query string false "date"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /user/{user_id}/operations [get]
func (u *userRoutes) getUserOperations(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date := r.URL.Query().Get("date")

	var operations []string
	var queryErr error

	if date == "" {
		operations, queryErr = u.userService.GetUserOperations(r.Context(), userId)
	} else {
		operations, queryErr = u.userService.GetUserOperationsByMonth(r.Context(), userId, date)
	}

	if queryErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(http.StatusOK)
	for _, operation := range operations {
		w.Write([]byte(operation + "\n"))
	}
}

// @Summary Get user operations report link
// @Description Returns a link to a report with a list of operations for the given user
// @Tags User
// @Param user_id path int true "user_id"
// @Param date query string false "date"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /user/{user_id}/operations/report-link [get]
func (u *userRoutes) getUserOperationsYandex(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "user_id")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date := r.URL.Query().Get("date")

	var operations []string
	var queryErr error

	if date == "" {
		operations, queryErr = u.userService.GetUserOperations(r.Context(), userId)
	} else {
		operations, queryErr = u.userService.GetUserOperationsByMonth(r.Context(), userId, date)
	}

	if queryErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var fileName string

	if date == "" {
		fileName = fmt.Sprintf("%d.csv", userId)
	} else {
		fileName = fmt.Sprintf("%d_%s.csv", userId, date)
	}

	url, err := u.userService.UploadAndReturnDownloadURL(r.Context(), fileName, operations)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}
