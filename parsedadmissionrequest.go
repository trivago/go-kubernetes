package kubernetes

import (
	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type parseResult int
type dataField int

const (
	parsedOk        = parseResult(iota)
	parsedWithError = parseResult(iota)
	parsedAsIgnore  = parseResult(iota)
)

const (
	dataFieldNew = dataField(iota)
	dataFieldOld = dataField(iota)
)

type ParsedAdmissionRequest struct {
	name      string
	namespace string

	gvr schema.GroupVersionResource

	incomingRaw *runtime.RawExtension
	incomingObj NamespacedObject
	existingRaw *runtime.RawExtension
	existingObj NamespacedObject
}

// ParseRequest converts an kubernetes AdmissionRequest into a parsed request.
func ParseRequest(req *admission.AdmissionRequest) ParsedAdmissionRequest {
	return ParsedAdmissionRequest{
		name:        req.Name,
		namespace:   req.Namespace,
		incomingRaw: &req.Object,
		existingRaw: &req.OldObject,
		gvr:         schema.GroupVersionResource(*req.RequestResource),
	}
}

// NewParsedAdmissionRequest creates a new ParsedAdmissionRequest from a given
// resources. This can be used to simulate AdmissionRequests.
func NewParsedAdmissionRequest(gvr schema.GroupVersionResource, name, namespace string, new, old NamespacedObject) ParsedAdmissionRequest {
	return ParsedAdmissionRequest{
		name:        name,
		namespace:   namespace,
		incomingObj: new,
		existingObj: old,
		gvr:         gvr,
	}
}

// GetName returns the name assigned to the admission request.
// This should be equal to GetNewObject().GetName()
func (p *ParsedAdmissionRequest) GetName() string {
	return p.name
}

// GetNamespace returns the namespace assigned to the admission request.
func (p *ParsedAdmissionRequest) GetNamespace() string {
	return p.namespace
}

// GetGroupVersionResource returns the GroupVersionResource assigned to this
// request.
func (p *ParsedAdmissionRequest) GetGroupVersionResource() schema.GroupVersionResource {
	return p.gvr
}

// Returns the incoming object raw json string
func (p *ParsedAdmissionRequest) GetIncomingJSON() []byte {
	if p.incomingRaw == nil {
		return []byte{}
	}
	return p.incomingRaw.Raw
}

// GetIncomingObject returns the object to be placed on the cluster.
// This object is only available on Create and Update requests.
func (p *ParsedAdmissionRequest) GetIncomingObject() (NamespacedObject, error) {
	if len(p.incomingObj) == 0 {
		var err error
		p.incomingObj, err = NamespacedObjectFromRaw(p.incomingRaw)
		if err != nil {
			return nil, err
		}
	}

	return p.incomingObj, nil
}

// GetExistingObject returns the object existing on the cluster.
// This object is only available on Delete and Update requests.
func (p *ParsedAdmissionRequest) GetExistingObject() (NamespacedObject, error) {
	if len(p.existingObj) == 0 {
		var err error
		p.existingObj, err = NamespacedObjectFromRaw(p.existingRaw)
		if err != nil {
			return nil, err
		}
	}

	return p.existingObj, nil
}
