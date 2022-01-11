package handler

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"grader/internal/app/panel/model"
	"grader/internal/app/panel/pkg/auth"
	"grader/internal/app/panel/storage"
	"grader/pkg/apperr"
	"grader/pkg/aws"
	"grader/pkg/httputil"
	"grader/pkg/layout"
	"grader/pkg/logger"
	"grader/pkg/queue"
	"mime/multipart"
	"net/http"
)

type SubmissionHandler struct {
	layout      *layout.Layout
	users       storage.UserRepository
	assessments storage.AssessmentRepository
	submissions storage.SubmissionRepository
	s3          *aws.S3
	topic       queue.Topic
}

func NewSubmitHandler(
	l *layout.Layout,
	s3 *aws.S3,
	q queue.Queue,
	u storage.UserRepository,
	a storage.AssessmentRepository,
	s storage.SubmissionRepository,
) (*SubmissionHandler, error) {
	const topicName = "grader-submissions"
	t, err := q.Topic(topicName)
	if err != nil {
		return nil, err
	}

	return &SubmissionHandler{
		layout:      l,
		users:       u,
		assessments: a,
		submissions: s,
		s3:          s3,
		topic:       t,
	}, nil
}

func (h *SubmissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := logger.Ctx(ctx)

	user, err := auth.UserFromContext(ctx)
	if err != nil {
		http.Error(w, apperr.ErrForbidden.Error(), http.StatusForbidden)
	}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		http.Error(w, "ID not found", http.StatusNotFound)
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Bad ID", http.StatusNotFound)
		return
	}

	as, err := h.assessments.Read(ctx, id)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Missing ID", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		h.layout.RenderView(w, r, "template/app/views/submit/create.gohtml", nil)
		return
	}

	if err := r.ParseMultipartForm(5 * 1024 * 1025); err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Form parse error", http.StatusInternalServerError)
		return
	}
	uploadData, _, err := r.FormFile("submission_file")
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "File upload error", http.StatusInternalServerError)
		return
	}
	defer func(uploadData multipart.File) {
		_ = uploadData.Close()
	}(uploadData)

	mType, err := mimetype.DetectReader(uploadData)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Unable to detect mime type", http.StatusInternalServerError)
		return
	}

	submissionID := uuid.New()

	fileName := fmt.Sprintf("%s/%s", submissionID.String(), as.FileName)

	if err := h.s3.Put(uploadData, fileName, mType.String(), user.ID.String()); err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		return
	}

	fileURL, err := h.s3.GetLink(fileName)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		return
	}

	l.Debug().Str("download-url", fileURL).Msg("Got download link")

	m := &model.Submission{
		UserID:       user.ID,
		AssessmentID: as.ID,
		FileName:     as.FileName,
		FileURL:      fileURL,
	}

	if _, err := h.submissions.Create(ctx, m); err != nil {
		l.Error().Err(err).Send()
		httputil.WriteError(w, apperr.ErrInternal, http.StatusInternalServerError)
		return
	}

	if err := h.topic.Publish(m); err != nil {
		l.Error().Err(err).Send()
		httputil.WriteError(w, apperr.ErrInternal, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app/user/submissions", http.StatusFound)
}
