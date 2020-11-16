package pod

import (
	"encoding/json"
	"fmt"
	"k8s-sync/pkg/constant"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

func IsPodReady(pod *v1.Pod) bool {
	if pod == nil {
		return false
	}
	if len(pod.Status.Conditions) < 4 {
		return false
	}
	for k, v := range pod.Status.Conditions {
		if k != 0 && v.Status != "True" {
			return false
		}
	}
	return true
}

//FetchLabelValue fetch label value from pod
func FetchLabelValue(pod *v1.Pod, labelName string, defaultValue string) string {
	if len(labelName) == 0 {
		return defaultValue
	}
	if labelValue, ok := pod.Labels[labelName]; ok {
		return labelValue
	}
	return defaultValue
}

func FetchEnvVar(pod *v1.Pod, envName, defaultValue string) string {
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if strings.Compare(envName, env.Name) == 0 {
				return env.Value
			}
		}
	}
	return defaultValue

}
func FetchEnvVarInt(pod *v1.Pod, envName string, defaultValue int) int {
	v := FetchEnvVar(pod, envName, "")
	if len(v) == 0 {
		return defaultValue
	}
	intValue, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func FetchPodAnnotation(pod *v1.Pod, annotationKey string) string {
	annotations := pod.Annotations
	if annotations == nil || len(annotations) == 0 {
		return ""
	}
	if value, ok := annotations[annotationKey]; ok {
		return value
	}
	return ""
}

func FetchPodIp(pod *v1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Status.PodIP
}
func FetchPodBillIdInt64(pod *v1.Pod, defaultValue int64) int64 {
	// parse billId
	billIdStr := FetchLabelValue(pod, constant.LABEL_BILL_ID, "")
	if len(billIdStr) == 0 {
		billIdStr = FetchEnvVar(pod, constant.EnvPublishBillId, "")
	}
	if len(billIdStr) == 0 {
		return defaultValue
	}
	billId, err := strconv.ParseInt(billIdStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return billId
}

func FetchPodIsolation(pod *v1.Pod, defaultValue string) string {
	// parse billId
	isolation := FetchLabelValue(pod, constant.LABEL_ISOLATION, "")
	if len(isolation) == 0 {
		isolation = FetchEnvVar(pod, constant.EnvIsolation, "")
	}
	if len(isolation) == 0 {
		return defaultValue
	}
	return isolation
}

func PatchPodLabels(kubeClient *kubernetes.Clientset, pod *v1.Pod, labels map[string]string) error {
	opLabel := "replace"
	if len(pod.Labels) == 0 {
		opLabel = "add"
	}
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}

	for k, v := range labels {
		pod.Labels[k] = v
	}
	patchLabelsPayloadTemplate :=
		`[{
        "op": "%s",
        "path": "/metadata/labels",
        "value": %s
    }]`
	rawLabels, _ := json.Marshal(pod.Labels)
	patchLabelsPayload := fmt.Sprintf(patchLabelsPayloadTemplate, opLabel, rawLabels)
	if _, err := kubeClient.CoreV1().Pods("default").Patch(pod.Name, types.JSONPatchType, []byte(patchLabelsPayload)); err != nil {
		return err
	}
	return nil
}

//PatchPodAnnotations just for log pod online time info
func PatchPodAnnotations(kubeClient *kubernetes.Clientset, pod *v1.Pod, annotations map[string]string) error {
	op := "replace"
	if len(pod.Annotations) == 0 {
		op = "add"
	}
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	for k, v := range annotations {
		if _, ok := pod.Annotations[k]; ok {
			//log.Debug().Msgf("Annotation has exist in pod k:% v:%s",k,pod.Annotations[k])
			return nil
		}
		pod.Annotations[k] = v
	}

	patchPayloadTemplate :=
		`[{
        "op": "%s",
        "path": "/metadata/annotations",
        "value": %s
    }]`

	raw, _ := json.Marshal(pod.Annotations)
	patchPayload := fmt.Sprintf(patchPayloadTemplate, op, raw)
	if _, err := kubeClient.CoreV1().Pods("default").Patch(pod.Name, types.JSONPatchType, []byte(patchPayload)); err != nil {
		return err
	}
	return nil
}
