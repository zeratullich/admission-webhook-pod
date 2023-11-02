package webhook

import (
	"admission-webhook-pod/options"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wI2L/jsondiff"
	"gopkg.in/yaml.v2"
	admissionv1 "k8s.io/api/admission/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

const (
	admissionWebhookAnnotationStatusKey = "admission-webhook-pod.zeratullich.com/status"
	admissionWebhookLabelMutateKey      = "admission-webhook-pod.zeratullich.com/app"
)

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}

	addLabels = map[string]string{
		admissionWebhookLabelMutateKey: "true",
	}
	addAnnotations = map[string]string{
		admissionWebhookAnnotationStatusKey: "mutated",
	}
)

type WebhookServer struct {
	Server *http.Server
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Serve method for webhook server
func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		logString := "empty body"
		log.Warnln(logString)
		http.Error(w, logString, http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logString := fmt.Sprintf("Content-Type=%s, expect `application/json`", contentType)
		log.Warnln(logString)
		http.Error(w, logString, http.StatusUnsupportedMediaType)
		return
	}

	if obj, gvk, err := deserializer.Decode(body, nil, &admissionv1.AdmissionReview{}); err != nil {
		logString := fmt.Sprintf("Request could not be decoded: %v ", err)
		log.Errorln(logString)
		http.Error(w, logString, http.StatusBadRequest)
		return
	} else {
		if r.URL.Path == options.MutatePath {
			requestedAdmissionReview, ok := obj.(*admissionv1.AdmissionReview)
			if !ok {
				logString := fmt.Sprintf("Expected v1.AdmissionReview but got: %T", obj)
				log.Errorln(logString)
				http.Error(w, logString, http.StatusBadRequest)
				return
			}

			admissionResponse := whsvr.mutate(requestedAdmissionReview)

			responseAdmissionReview := &admissionv1.AdmissionReview{}
			responseAdmissionReview.SetGroupVersionKind(*gvk)

			if admissionResponse != nil {
				responseAdmissionReview.Response = admissionResponse
				if requestedAdmissionReview.Request != nil {
					responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
				}
			}

			resp, err := json.Marshal(responseAdmissionReview)
			if err != nil {
				logString := fmt.Sprintf("Can't encode response: %v", err)
				log.Errorln(logString)
				http.Error(w, logString, http.StatusInternalServerError)
			}

			log.Infoln("Ready to write response...")
			if _, err := w.Write(resp); err != nil {
				logString := fmt.Sprintf("Can't write response: %v", err)
				log.Errorln(logString)
				http.Error(w, logString, http.StatusInternalServerError)
			}

			dateTime := time.Now().In(time.FixedZone("GMT", 8*3600)).Format("2006-01-02 15:04:05")
			logString := fmt.Sprintf("======Admission has written to response already at %s======", dateTime)
			log.Infoln(logString)
		}
	}
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	pod := corev1.Pod{}

	switch req.Kind.Kind {
	case "Pod":
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			log.Errorf("Can't unmarshal raw object: %v", err)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
	default:
		msg := fmt.Sprintf("Not support for this Kind of resource %v", req.Kind.Kind)
		log.Warnln(msg)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: msg,
			},
		}
	}

	var name string
	if pod.Name != "" {
		name = pod.Name
	} else {
		name = pod.GenerateName
	}

	log.Infof("======Begin Admission for Namespace=[%v], Kind=[%v], Name=[%v]======", req.Namespace, req.Kind.Kind, name)
	log.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v PatchOperation=%v UserInfo=%v",
		req.Kind.Kind, req.Namespace, name, req.UID, req.Operation, req.UserInfo)

	// pod.Namespace is empty value , so must set namespace value is req.Namespace
	if !mutationRequired(ignoredNamespaces, &pod.ObjectMeta, req.Namespace) {
		log.Infof("Skip validation for %s/%s , due to policy check", req.Namespace, name)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := createPatch(&pod, addAnnotations, addLabels)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Debugf("AdmissionResponse: patch=%v", string(patchBytes))
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}

}

// skip some namespaces or limitary tags
func admissionRequired(ignoredList []string, metadata *metav1.ObjectMeta, namespace string) bool {
	for _, ns := range ignoredList {
		if namespace == ns {
			log.Infof("Skip validation for %+v , it's in special namespace: %v", metadata.OwnerReferences, namespace)
			return false
		}
	}

	var name string
	if metadata.Name != "" {
		name = metadata.Name
	} else {
		name = metadata.GenerateName
	}

	annotations := metadata.GetAnnotations()
	switch strings.ToLower(annotations[admissionWebhookAnnotationStatusKey]) {
	case "n", "no", "false", "off":
		log.Infof("Skip validation for pod: %s/%s , it's in special annotations: %s:%s", namespace, name, admissionWebhookAnnotationStatusKey, annotations[admissionWebhookAnnotationStatusKey])
		return false
	}

	labels := metadata.GetLabels()
	switch strings.ToLower(labels[admissionWebhookLabelMutateKey]) {
	case "n", "no", "false", "off":
		log.Infof("Skip validation for pod: %s/%s , it's in special labels: %s:%s", namespace, name, admissionWebhookLabelMutateKey, labels[admissionWebhookLabelMutateKey])
		return false
	}

	return true
}

// check whether the target resoured need to be mutated
func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta, namespace string) bool {

	var name string
	if metadata.Name != "" {
		name = metadata.Name
	} else {
		name = metadata.GenerateName
	}

	required := admissionRequired(ignoredList, metadata, namespace)
	if !required {
		return required
	}

	annotations := metadata.GetAnnotations()
	labels := metadata.GetLabels()

	if strings.ToLower(annotations[admissionWebhookAnnotationStatusKey]) == "mutated" || strings.ToLower(labels[admissionWebhookLabelMutateKey]) == "true" {
		required = false
	}
	log.Infof("Mutation policy for pod => %s/%s : status: %q required: %v", namespace, name, annotations[admissionWebhookAnnotationStatusKey], required)
	return required
}

// create mutation patch for resoures
func createPatch(pod *corev1.Pod, addAnnotations map[string]string, addLabels map[string]string) ([]byte, error) {
	var patches []patchOperation
	objectMeta := pod.ObjectMeta
	labels := objectMeta.Labels
	annotations := objectMeta.Annotations
	labelsPatch := updateLabels(labels, addLabels)
	annotationsPatch := updateAnnotations(annotations, addAnnotations)
	containersPatch := updateContainers(addContainer, pod)

	patches = append(patches, labelsPatch...)
	patches = append(patches, annotationsPatch...)
	patches = append(patches, containersPatch...)

	patchYAML, err := yaml.Marshal(patches)
	if err != nil {
		log.Errorf("Patch To PatchYAML Failure: %s", err)
	}
	log.Debugf("The modification content is as follows: %s", string(patchYAML))
	return json.Marshal(patches)
}

// update some annotations
func updateAnnotations(target map[string]string, added map[string]string) (patch []patchOperation) {
	if target == nil {
		target = make(map[string]string)
	}

	for key, value := range added {
		target[key] = value
	}

	patch = append(patch, patchOperation{
		Op:    "replace",
		Path:  "/metadata/annotations",
		Value: target,
	})
	return patch
}

// update some labels
func updateLabels(target map[string]string, added map[string]string) (patch []patchOperation) {
	if target == nil {
		target = make(map[string]string)
	}

	for key, value := range added {
		target[key] = value
	}

	patch = append(patch, patchOperation{
		Op:    "replace",
		Path:  "/metadata/labels",
		Value: target,
	})
	return patch
}

// side-car
var addContainer = []corev1.Container{
	{
		Name:    "side-car",
		Image:   "busybox",
		Command: []string{"/bin/sleep", "infinity"},
	},
}

// add some containers for pod(spec.containers)
func updateContainers(addContainer []corev1.Container, pod *corev1.Pod) (patch []patchOperation) {
	currentPod := pod.DeepCopy()
	containers := currentPod.Spec.Containers
	containers = append(containers, addContainer...)
	currentPod.Spec.Containers = containers
	diffPatch, err := jsondiff.Compare(pod, currentPod)
	if err != nil {
		log.Error(err)
	}
	for _, v := range diffPatch {
		addPatch := patchOperation{
			Op:    v.Type,
			Value: v.Value,
			Path:  v.Path,
		}
		patch = append(patch, addPatch)
	}
	return patch
}
