// Inspired by
// https://github.com/douglasmakey/admissionkubernetes

package kubernetes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidationFunc callback function prototype for hooks
type ValidationFunc func(req ParsedAdmissionRequest) ValidationResult

// AdmissionRequestHook is a helper struct to automaticall map admission
// operations to functions.
type AdmissionRequestHook struct {
	Create ValidationFunc
	Delete ValidationFunc
	Update ValidationFunc
}

// Call runs the correct callback per requested operation.
// If an operation does not have a callback registered, an error is reported,
// but the request is reported as validated.
func (h AdmissionRequestHook) Call(req *admission.AdmissionRequest) ValidationResult {
	callback := ValidationFunc(nil)

	switch req.Operation {
	case admission.Create:
		callback = h.Create
	case admission.Update:
		callback = h.Update
	case admission.Delete:
		callback = h.Delete
	default:
		log.Error().Msgf("unknown admission operation: %s", req.Operation)
		return ValidationOk
	}

	if callback == nil {
		log.Error().Msgf("operation %s has no callback set", req.Operation)
		return ValidationOk
	}

	// TODO: create parse request here
	parsed := ParseRequest(req)
	return callback(parsed)
}

// Handle reads an admission request, calls the corresponding hook and builds
// the correct response object.
func (h AdmissionRequestHook) Handle(ctx *gin.Context) {
	admissionResponse := admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}

	// Parse admission review from body. If this failes we report back a malformed request.
	review := new(admission.AdmissionReview)
	if err := ctx.BindJSON(review); err != nil {
		log.Error().Err(err).Msg("failed to decode admission review")
		ctx.Status(http.StatusBadRequest)
		return
	}

	// Always return ok for dry runs
	if review.Request.DryRun != nil && *review.Request.DryRun {
		log.Info().Msg("ignored admission request for dry run")
		admissionResponse.Response = NewOkResponse(review.Request)
		ctx.AsciiJSON(http.StatusOK, admissionResponse)
		return
	}

	// Call the review handler
	result := h.Call(review.Request)
	admissionResponse.Response = result.ToResponse(review.Request)

	// Return response as proper JSON
	ctx.AsciiJSON(http.StatusOK, admissionResponse)
}
