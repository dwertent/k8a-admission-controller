package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"net/http"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAWebhookHandleRequest Webhook handler
func CAWebhookHandleRequest(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorf("In CAWebhookHandleRequest, recover: %v", err)
		}
	}()

	var admissionResponse *admissionv1.AdmissionResponse
	admissionReview, err := parseRequest(r)
	if err != nil {
		e := fmt.Errorf("failed to read APIServer request, reason: %s", err.Error())
		glog.Errorf(e.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	admissionResponse = handleWorkloadRequest(admissionReview, &w, r)
	if admissionResponse == nil {
		glog.Errorf("admissionResponse is nil")
		admissionResponse = &admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Message: "admissionResponse is nil",
			},
		}
	}

	admissionReview.Response = admissionResponse
	if admissionReview.Request != nil {
		admissionReview.Response.UID = admissionReview.Request.UID
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}
	// glog.Infof("response: %s", string(resp))
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func handleWorkloadRequest(ar *admissionv1.AdmissionReview, w *http.ResponseWriter, r *http.Request) *admissionv1.AdmissionResponse {
	admissionResponse := &admissionv1.AdmissionResponse{}
	var inter interface{}
	var err error
	if err != nil {
		glog.Errorf("%v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if inter == nil {
			admissionResponse = &admissionv1.AdmissionResponse{
				Allowed: true,
			}
		} else if str, ok := inter.(string); ok {
			admissionResponse = &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: str,
				},
			}
			// cautils.SendSafeModeReport(carh.GetWlid(), carh.GetPod().GetName(), carh.GetReporter().GetJobID(), str, 1)
		} else if patchBytes, ok := inter.([]byte); ok {
			admissionResponse = &admissionv1.AdmissionResponse{
				Allowed: true,
				Patch:   patchBytes,
				PatchType: func() *admissionv1.PatchType {
					pt := admissionv1.PatchTypeJSONPatch
					return &pt
				}(),
			}
		}
	}

	return admissionResponse
}

var tmp = 0

type Kind struct {
	Kind string `json:"kind"`
}
type Resource struct {
	Resource string `json:"resource"`
}
type Request struct {
	Kind      Kind     `json:"kind"`
	Resource  Resource `json:"resource"`
	Namespace string   `json:"namespace"`
	Operation string   `json:"operation"`
}
type Req struct {
	Request Request `json:"request"`
}

func parseRequest(r *http.Request) (*admissionv1.AdmissionReview, error) {
	// read message received from k8sAPI
	var body []byte

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}

	if tmp <= 1 {
		fmt.Printf("Request start ===========================================================\n")
		fmt.Printf("%s\n", body)
		tmp += 1
		fmt.Printf("Request end *************************************************************\n")
	}

	// Verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("Content-Type=%s, expect application/json", contentType)
	}

	ar := &admissionv1.AdmissionReview{}
	if err := json.Unmarshal(body, ar); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	if ar.Request.Kind.Kind != "Lease" && ar.Request.Kind.Kind != "Endpoints" {
		fmt.Printf("Request start ===========================================================\n")
		fmt.Printf("%s\n", body)
		fmt.Printf("Request end *************************************************************\n")
	}
	if ar.Request.Operation != "UPDATE" {
		fmt.Printf("Request start ===========================================================\n")
		fmt.Printf("K8s Obj: %s, Req Kind %s, NS: %s, Operation: %s, Resource: %s\n", ar.Kind, ar.Request.Kind.Kind, ar.Request.Namespace, ar.Request.Operation, ar.Request.Resource.Resource)
		fmt.Printf("UserName: %s\n", ar.Request.UserInfo.Username)
		fmt.Printf("Request end *************************************************************\n")
	}
	reflect.TypeOf(ar.Request.Object)
	//kind := ar.Request.Object.Object
	if ar.Request.Object.Object != nil {
		//if ar.Request.Object.Object.GetObjectKind() != nil {
		fmt.Printf("Object Kind: %v", ar.Request.Object.Object.GetObjectKind())
	}

	if ar.Request.OldObject.Object != nil {
		//if ar.Request.OldObject.Object.GetObjectKind() != nil {
		fmt.Printf("Old Object Kind: %v", ar.Request.OldObject.Object.GetObjectKind())
	}

	//var req map[string][]interface{}
	//if err := json.Unmarshal([]byte(body), t); err != nil {
	//	fmt.Println(err.Error())
	//} else {
	//	fmt.Println(t.Request.Namespace, t.Request.Operation, t.Request.Resource.Resource, t.Request.Kind.Kind)
	//}

	//fmt.Printf("%s\n", body[40:119])

	return ar, nil
}
