package main

import (
	"embed"
	"errors"
	"net/http"

	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"github.com/godruoyi/go-snowflake"
	"github.com/ip812/go-template/config"
	"github.com/ip812/go-template/database"
	"github.com/ip812/go-template/logger"
	"github.com/ip812/go-template/status"
	"github.com/ip812/go-template/templates/components"
	"github.com/ip812/go-template/templates/views"
	"github.com/ip812/go-template/utils"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

//go:embed static
var staticFS embed.FS

type Handler struct {
	config        *config.Config
	formDecoder   *form.Decoder
	formValidator *validator.Validate
	tracer        oteltrace.Tracer
	log           logger.Logger

	db DBWrapper
}

func (hnd *Handler) StaticFiles() http.Handler {
	if hnd.config.App.Env == config.Local {
		hnd.log.Info("serving static files from local directory")
		return http.StripPrefix("/static", http.FileServer(http.Dir("static")))
	}

	hnd.log.Info("serving static files from embedded FS")
	return http.StripPrefix("/", http.FileServer(http.FS(staticFS)))
}

func (hnd *Handler) LandingPageView(w http.ResponseWriter, r *http.Request) {
	utils.Render(w, r, views.LandingPage())
}

func (hnd *Handler) AddEmailToMailingList(w http.ResponseWriter, r *http.Request) error {
	queries, err := hnd.db.Queries()
	if err != nil {
		status.AddToast(w, status.ErrorInternalServerError(err))
		return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
	}

	err = r.ParseForm()
	if err != nil {
		status.AddToast(w, status.ErrorInternalServerError(status.ErrParsingFrom))
		return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
	}
	var props components.MailingListFormProps
	err = hnd.formDecoder.Decode(&props, r.Form)
	if err != nil {
		status.AddToast(w, status.ErrorInternalServerError(status.ErrDecodingForm))
		return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
	}

	_, span := hnd.tracer.Start(
		r.Context(),
		"AddEmailToMailingList",
		oteltrace.WithAttributes(attribute.String("email", props.Email)),
	)
	defer span.End()

	err = hnd.formValidator.Struct(props)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			status.AddToast(w, status.ErrorInternalServerError(status.ErrFailedtoValidateRequest))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
		}

		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Email":
				if err.Tag() == "required" {
					status.AddToast(w, status.WarningStatusBadRequest(status.WarnEmailIsRequred))
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{Email: props.Email}))
				} else if err.Tag() == "email" {
					status.AddToast(w, status.WarningStatusBadRequest(status.WarnInvalidEmailFormat))
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{Email: props.Email}))
				}
			}
		}
	}

	output, err := queries.AddEmailToMailingList(
		r.Context(),
		database.AddEmailToMailingListParams{
			ID:    int64(snowflake.ID()),
			Email: props.Email,
		},
	)
	if err != nil {
		var pgErr *pq.Error
		ok := errors.As(err, &pgErr)
		if ok {
			if pgErr.Code == "23505" {
				status.AddToast(w, status.WarningStatusBadRequest(status.WarnEmailAlreadyExists))
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
			}
		}

		status.AddToast(w, status.ErrorInternalServerError(status.ErrFailedToAddEmailToMailingList))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
	}
	hnd.log.Info("email %s was added to the mailing list", output.Email)

	status.AddToast(w, status.SuccessStatusCreated(status.SuccEmailAddedToMailingList))
	return utils.Render(w, r, components.MailingListForm(components.MailingListFormProps{}))
}

func (hnd *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

func (hnd *Handler) LandingPageRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/p/public/home", http.StatusFound)
}
