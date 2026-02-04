package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// EDIT THIS FILE: webhook defaulting/validation for JsonServer.

//+kubebuilder:webhook:path=/mutate-example-v1-jsonserver,mutating=true,failurePolicy=fail,sideEffects=NoneOnDryRun,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=mjsonserver.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:verbs=create;update,path=/validate-example-v1-jsonserver,mutating=false,failurePolicy=fail,sideEffects=NoneOnDryRun,groups=example.com,resources=jsonservers,versions=v1,name=vjsonserver.kb.io,admissionReviewVersions=v1

var jsonserverLog = log.Log.WithName("jsonserver-resource")

// SetupWebhookWithManager registers mutating and validating webhook handlers
// directly on the manager webhook server so the admission paths declared by
// the kubebuilder markers are actually served at runtime.
func (r *JsonServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	server := mgr.GetWebhookServer()

	// Mutating webhook: apply defaults
	server.Register("/mutate-example-v1-jsonserver", &admission.Webhook{Handler: &jsonServerMutator{}})

	// Validating webhook: run ValidateCreate/ValidateUpdate
	server.Register("/validate-example-v1-jsonserver", &admission.Webhook{Handler: &jsonServerValidator{}})

	return nil
}

// jsonServerMutator applies defaults by calling the type's Default method
type jsonServerMutator struct{}

func (m *jsonServerMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	var js JsonServer
	if err := json.Unmarshal(req.AdmissionRequest.Object.Raw, &js); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Apply defaults
	js.Default(ctx)

	mod, err := json.Marshal(&js)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, mod)
}

// jsonServerValidator validates create/update operations using the type's
// ValidateCreate/ValidateUpdate methods.
type jsonServerValidator struct{}

func (v *jsonServerValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	var js JsonServer
	if err := json.Unmarshal(req.AdmissionRequest.Object.Raw, &js); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	switch req.AdmissionRequest.Operation {
	case admissionv1.Create:
		if err := js.ValidateCreate(ctx); err != nil {
			return admission.Denied(err.Error())
		}
		return admission.Allowed("create validated")
	case admissionv1.Update:
		var old JsonServer
		if err := json.Unmarshal(req.AdmissionRequest.OldObject.Raw, &old); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if err := js.ValidateUpdate(ctx, &old); err != nil {
			return admission.Denied(err.Error())
		}
		return admission.Allowed("update validated")
	default:
		return admission.Allowed("operation allowed")
	}
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *JsonServer) Default(ctx context.Context) {
	jsonserverLog.Info("default", "name", r.Name)
	// Default replicas to 1 if not provided
	if r.Spec.Replicas == nil {
		def := int32(1)
		r.Spec.Replicas = &def
	}
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateCreate(ctx context.Context) error {
	jsonserverLog.Info("validate create", "name", r.Name)
	// Name must start with "app-"
	if !strings.HasPrefix(r.Name, "app-") {
		return fmt.Errorf("name must start with app-")
	}

	// Validate jsonConfig is valid JSON (must be an object)
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(r.Spec.JsonConfig), &js); err != nil {
		return fmt.Errorf("Error: spec.jsonConfig is not a valid json object")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateUpdate(ctx context.Context, old runtime.Object) error {
	jsonserverLog.Info("validate update", "name", r.Name)

	// Name must start with "app-"
	if !strings.HasPrefix(r.Name, "app-") {
		return fmt.Errorf("name must start with app-")
	}

	// Validate jsonConfig is valid JSON (must be an object)
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(r.Spec.JsonConfig), &js); err != nil {
		return fmt.Errorf("Error: spec.jsonConfig is not a valid json object")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateDelete(ctx context.Context) error {
	jsonserverLog.Info("validate delete", "name", r.Name)

	// Add deletion validation if needed (most controllers permit deletes).
	return nil
}
