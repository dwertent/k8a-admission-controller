package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Mock() string {
	s := `{
		"kind": "AdmissionReview",
		"apiVersion": "admission.k8s.io/v1beta1",
		"request": {
		  "uid": "0480097d-85c0-44d4-b231-c3b618427bdf",
		  "kind": {
			"group": "",
			"version": "v1",
			"kind": "Event"
		  },
		  "resource": {
			"group": "",
			"version": "v1",
			"resource": "events"
		  },
		  "requestKind": {
			"group": "",
			"version": "v1",
			"kind": "Event"
		  },
		  "requestResource": {
			"group": "",
			"version": "v1",
			"resource": "events"
		  },
		  "name": "webhook-64c5cb5cf7-5hm7q.16e2bba77b6dab15",
		  "namespace": "test",
		  "operation": "CREATE",
		  "userInfo": {
			"username": "system:node:minikube",
			"groups": [
			  "system:nodes",
			  "system:authenticated"
			]
		  },
		  "object": {
			"kind": "Event",
			"apiVersion": "v1",
			"metadata": {
			  "name": "webhook-64c5cb5cf7-5hm7q.16e2bba77b6dab15",
			  "namespace": "test",
			  "uid": "80ecad9e-8270-4cb9-94f5-19183a3b0d59",
			  "creationTimestamp": "2022-04-04T15:39:02Z",
			  "managedFields": [
				{
				  "manager": "kubelet",
				  "operation": "Update",
				  "apiVersion": "v1",
				  "time": "2022-04-04T15:39:02Z",
				  "fieldsType": "FieldsV1",
				  "fieldsV1": {
					"f:count": {},
					"f:firstTimestamp": {},
					"f:involvedObject": {},
					"f:lastTimestamp": {},
					"f:message": {},
					"f:reason": {},
					"f:source": {
					  "f:component": {},
					  "f:host": {}
					},
					"f:type": {}
				  }
				}
			  ]
			},
			"involvedObject": {
			  "kind": "Pod",
			  "namespace": "test",
			  "name": "webhook-64c5cb5cf7-5hm7q",
			  "uid": "cd0612eb-07a6-4c1e-9911-7d285eec9f5d",
			  "apiVersion": "v1",
			  "resourceVersion": "74652",
			  "fieldPath": "spec.containers{webhook}"
			},
			"reason": "Created",
			"message": "Created container webhook",
			"source": {
			  "component": "kubelet",
			  "host": "minikube"
			},
			"firstTimestamp": "2022-04-04T15:39:01Z",
			"lastTimestamp": "2022-04-04T15:39:01Z",
			"count": 1,
			"type": "Normal",
			"eventTime": null,
			"reportingComponent": "",
			"reportingInstance": ""
		  },
		  "oldObject": null,
		  "dryRun": false,
		  "options": {
			"kind": "CreateOptions",
			"apiVersion": "meta.k8s.io/v1"
		  }
		}
	  }
	  `

	return s
}
func TestParse(t *testing.T) {
	a := &Req{}
	b := []byte(Mock())

	assert.NoError(t, json.Unmarshal(b, a))
	assert.Equal(t, "Event", a.Request.Kind.Kind)
}
