package v1

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/realPointer/segments/internal/service"
	"github.com/realPointer/segments/pkg/logger"
)

type segmentRoutes struct {
	segmentService service.Segment
}

func NewSegmentRouter(segmentService service.Segment, l logger.Interface) http.Handler {
	s := segmentRoutes{segmentService: segmentService}
	r := chi.NewRouter()

	r.Post("/{segmentName}", s.createSegment)
	r.Delete("/{segmentName}", s.deleteSegment)
	r.Get("/list", s.getSegments)

	return r
}

// @Summary Create segment
// @Description Creates a new segment with the given name
// @Tags Segment
// @Param segmentName path string true "segmentName"
// @Param auto query string false "auto"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /segment/{segmentName} [post]
func (s *segmentRoutes) createSegment(w http.ResponseWriter, r *http.Request) {
	segmentName := chi.URLParam(r, "segmentName")

	autoStr := r.URL.Query().Get("auto")

	var err error

	if autoStr == "" {
		err = s.segmentService.CreateSegment(r.Context(), segmentName)
	} else {
		percentage, parseErr := strconv.ParseFloat(autoStr, 64)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = s.segmentService.CreateSegmentAuto(r.Context(), segmentName, percentage)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary Delete segment
// @Description Deletes a segment with the given name
// @Tags Segment
// @Param segmentName path string true "segmentName"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /segment/{segmentName} [delete]
func (s *segmentRoutes) deleteSegment(w http.ResponseWriter, r *http.Request) {
	segmentName := chi.URLParam(r, "segmentName")
	err := s.segmentService.DeleteSegment(r.Context(), segmentName)
	if err != nil {
		log.Default().Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get segments
// @Description Returns a list of segments
// @Tags Segment
// @Success 200
// @Failure 500
// @Router /segment/list [get]
func (s *segmentRoutes) getSegments(w http.ResponseWriter, r *http.Request) {
	segments, err := s.segmentService.GetSegments(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, segments)
}
