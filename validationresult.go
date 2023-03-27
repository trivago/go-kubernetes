// Inspired by
// https://github.com/douglasmakey/admissionkubernetes

package kubernetes

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Result of a ValidationFunc.
type ValidationResult struct {
	// Ok holds the result of the validation
	Ok bool
	// Message can give additional context on the result
	Message string
	// Patches may hold modifications to be done on the validated object
	Patches []PatchOperation
}

var (
	ValidationOk     = ValidationResult{Ok: true}
	ValidationFailed = ValidationResult{Ok: false}
)

func NewErrorResponse(req *admission.AdmissionRequest, message string) *admission.AdmissionResponse {
	response := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: false,
		Result: &meta.Status{
			Message: message,
			Code:    503,
		},
	}

	return &response
}

func NewOkResponse(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	response := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	return &response
}

func (result ValidationResult) ToResponse(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	response := admission.AdmissionResponse{
		UID:     req.UID,
		Allowed: result.Ok,
	}

	if !result.Ok && len(result.Message) > 0 {
		response.Result = &meta.Status{
			Message: result.Message,
			Code:    503,
		}
	}

	if len(result.Patches) > 0 {
		if patchBytes, err := jsoniter.Marshal(result.Patches); err != nil {
			log.Error().Err(err).Msg("failed to encode patches")
		} else {
			patchType := admission.PatchTypeJSONPatch

			response.Patch = patchBytes
			response.PatchType = &patchType

			log.Debug().RawJSON("patch", patchBytes).Msg("patch provided")
		}
	} else {
		log.Debug().Msg("no patch provided")
	}

	return &response
}
