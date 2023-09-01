package cmd

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type EventResp struct {
	NAMESPACE     string    `json:"namespace"`
	NAME          string    `json:"name"`
	UID           types.UID `json:"uid"`
	LastTimestamp time.Time `json:"lastTimestamp"`
	Type          string    `json:"type"`
	Reason        string    `json:"reason"`
	Object        string    `json:"object"`
	Count         int32     `json:"count"`
	Message       string    `json:"message"`
}

type EventFlags struct {
	K8S
	Kafka
}

type K8S struct {
	KubeConfig string
}

type Kafka struct {
	AddrList string
	Topic    string
}
