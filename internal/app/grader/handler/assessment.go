package handler

import (
	"github.com/google/uuid"
	"grader/internal/app/grader/runner"
	"grader/pkg/httputil"
	"grader/pkg/workerpool"
	"net/http"
)

type AssessmentHandler struct {
	workers *workerpool.Pool
}

func NewAssessmentHandler(wp *workerpool.Pool) *AssessmentHandler {
	return &AssessmentHandler{
		workers: wp,
	}
}

type CheckAssessmentRequest struct {
	Assessment runner.Assessment `json:"assessment"`
}

type CheckAssessmentResponse struct {
	TaskID uuid.UUID `json:"task_id"`
}

type AssessmentFile struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required,url"`
	Path string `json:"path" validate:"required,file"`
}

/**
{
    "assessment": {
        "container_image": "hello-world",
        "part_id": "test",
        "postback_url": "http://localhost:8090/404",
        "files": [

        ]
    }
}
*/

func (h *AssessmentHandler) Check(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	//l := logger.Ctx(ctx)

	in := &CheckAssessmentRequest{}

	if err := httputil.ReadBody(r, in); err != nil {
		httputil.WriteError(w, err, http.StatusBadRequest)
		return
	}

	if !httputil.ValidateData(w, in) {
		return
	}

	in.Assessment.TaskID = uuid.New()

	h.workers.Run(runner.CheckAssessmentJob(in.Assessment))

	out := &CheckAssessmentResponse{
		TaskID: in.Assessment.TaskID,
	}

	httputil.WriteResponse(w, out, http.StatusAccepted)
}
