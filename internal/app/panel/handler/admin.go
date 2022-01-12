package handler

import (
	"grader/internal/app/panel/storage"
	"grader/internal/pkg/model"
	"grader/pkg/apperr"
	"grader/pkg/httputil"
	"grader/pkg/layout"
	"grader/pkg/logger"
	"net/http"
)

type AdminHandler struct {
	layout      *layout.Layout
	users       storage.UserRepository
	assessments storage.AssessmentRepository
}

func NewAdminHandler(l *layout.Layout, u storage.UserRepository, a storage.AssessmentRepository) *AdminHandler {
	return &AdminHandler{layout: l, users: u, assessments: a}
}

func (h *AdminHandler) AssessmentList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := logger.Ctx(ctx)

	l.Debug().Msg("Assessment List")

	models, err := h.assessments.All(ctx)
	if err != nil {
		l.Error().Err(err).Send()
		httputil.WriteError(w, err, http.StatusInternalServerError)
	}

	data := map[string]interface{}{
		"Models": models,
	}

	h.layout.RenderView(w, r, "template/app/views/admin/assessment_list.gohtml", data)
}

func (h *AdminHandler) AssessmentCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := logger.Ctx(ctx)

	if r.Method != http.MethodPost {
		h.layout.RenderView(w, r, "template/app/views/admin/assessment_create.gohtml", nil)
		return
	}

	in := &struct {
		PartID         string `validate:"required"`
		ContainerImage string `validate:"required"`
		Summary        string `validate:"required"`
		FileName       string `validate:"required"`
	}{
		r.FormValue("part_id"),
		r.FormValue("container_image"),
		r.FormValue("summary"),
		r.FormValue("file_name"),
	}

	if !httputil.ValidateData(w, in) {
		return
	}

	m := &model.Assessment{
		PartID:         in.PartID,
		ContainerImage: in.ContainerImage,
		Summary:        in.Summary,
		FileName:       in.FileName,
	}

	_, err := h.assessments.Create(ctx, m)
	if err != nil {
		l.Error().Err(err).Send()
		httputil.WriteError(w, apperr.ErrInternal, http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/app/admin/assessments", http.StatusFound)
}
