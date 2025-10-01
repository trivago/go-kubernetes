// Inspired by
// https://github.com/douglasmakey/admissionkubernetes

package kubernetes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
func (h AdmissionRequestHook) Call(req *admission.AdmissionRequest) (ValidationResult, error) {
	callback := ValidationFunc(nil)

	switch req.Operation {
	case admission.Create:
		callback = h.Create
	case admission.Update:
		callback = h.Update
	case admission.Delete:
		callback = h.Delete
	default:
		return ValidationOk, fmt.Errorf("unknown admission operation: %s", req.Operation)
	}

	if callback == nil {
		return ValidationOk, fmt.Errorf("operation %s has no callback set", req.Operation)
	}

	// TODO: create parse request here
	parsed := ParseRequest(req)
	return callback(parsed), nil
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
		_ = ctx.Error(errors.Wrapf(err, "failed to parse admission review"))
		ctx.Status(http.StatusBadRequest)
		return
	}

	// Always return ok for dry runs
	if review.Request.DryRun != nil && *review.Request.DryRun {
		admissionResponse.Response = NewOkResponse(review.Request)
		ctx.AsciiJSON(http.StatusOK, admissionResponse)
		return
	}

	// Call the review handler
	result, err := h.Call(review.Request)
	if err != nil {
		_ = ctx.Error(err)
	}

	// Convert the response
	admissionResponse.Response, err = result.ToResponse(review.Request)
	if err != nil {
		_ = ctx.Error(err)
	}

	// Return response as proper JSON
	ctx.AsciiJSON(http.StatusOK, admissionResponse)
}
