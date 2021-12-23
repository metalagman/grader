package handler

import (
	"github.com/google/uuid"
	"grader/internal/app/grader/runner"
	"grader/pkg/httputil"
	"grader/pkg/workerpool"
	"net/http"
)

type SubmissionHandler struct {
	workers *workerpool.Pool
}

func NewSubmissionHandler(wp *workerpool.Pool) *SubmissionHandler {
	return &SubmissionHandler{
		workers: wp,
	}
}

type CheckSubmissionRequest struct {
	Submission runner.Submission `json:"submission"`
}

type CheckSubmissionResponse struct {
	TaskID uuid.UUID `json:"task_id"`
}

/**
{
    "submission": {
        "container_image": "hello-world",
        "part_id": "test",
        "postback_url": "http://localhost:8090/404",
        "files": [

        ]
    }
}
*/

func (h *SubmissionHandler) Check(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	//l := logger.Ctx(ctx)

	in := &CheckSubmissionRequest{}

	if err := httputil.ReadBody(r, in); err != nil {
		httputil.WriteError(w, err, http.StatusBadRequest)
		return
	}

	if !httputil.ValidateData(w, in) {
		return
	}

	in.Submission.TaskID = uuid.New()

	h.workers.Run(runner.CheckSubmissionJob(in.Submission))

	out := &CheckSubmissionResponse{
		TaskID: in.Submission.TaskID,
	}

	httputil.WriteResponse(w, out, http.StatusAccepted)
}
