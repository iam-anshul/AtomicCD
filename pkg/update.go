package pkg

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func update(cluster *kubernetes.Clientset, targetContainerConfig *[]map[string]interface{}, namespace string, deploymenName string, deployType string) {
	containerSpecs := ""
	for i, config := range *targetContainerConfig {
		containerSpec := fmt.Sprintf(`{
		"name": "%s",
		"image": "%s:%s"
		}`, config["containername"], config["containerimage"], config["containertag"])

		containerSpecs += containerSpec

		if i < len((*targetContainerConfig))-1 {
			containerSpecs += ","
		}
	}

	patchPayload := fmt.Sprintf(`{
		"spec": {
			"template": {
				"spec": {
					"containers": [
					%s
					]
				}
			}
		}
	}`, containerSpecs)

	if deployType == "deployment" {
		_, err := cluster.AppsV1().Deployments(namespace).Patch(
			context.TODO(),
			deploymenName,
			types.StrategicMergePatchType,
			[]byte(patchPayload),
			metav1.PatchOptions{},
		)
		if err != nil {
			log.Fatal("Error occured in update func's patch call. Error is: ", err)
		}
	} else {
		_, err := cluster.AppsV1().StatefulSets(namespace).Patch(
			context.TODO(),
			deploymenName,
			types.StrategicMergePatchType,
			[]byte(patchPayload),
			metav1.PatchOptions{},
		)
		if err != nil {
			log.Fatal("Error occured in update func's patch call. Error is: ", err)
		}
	}
}

func updateInNamespace(cluster *kubernetes.Clientset, podInfo *[]map[string]string, deploymenName string, namespace string, deployType string) {
	containerSpecs := ""
	for i, containerInfo := range *podInfo {
		containerSpec := fmt.Sprintf(`{
		"name": "%s",
		"image": "%s:%s"
		}`, containerInfo["containerName"], containerInfo["containerImage"], containerInfo["containerTag"])

		containerSpecs += containerSpec

		if i < len((*podInfo))-1 {
			containerSpecs += ","
		}
	}

	patchPayload := fmt.Sprintf(`{
		"spec": {
			"template": {
				"spec": {
					"containers": [
					%s
					]
				}
			}
		}
	}`, containerSpecs)

	if deployType == "deployment" {
		_, err := cluster.AppsV1().Deployments(namespace).Patch(
			context.TODO(),
			deploymenName,
			types.StrategicMergePatchType,
			[]byte(patchPayload),
			metav1.PatchOptions{},
		)
		if err != nil {
			log.Fatal("Error occured in update func's patch call. Error is: ", err)
		}
	} else {
		_, err := cluster.AppsV1().StatefulSets(namespace).Patch(
			context.TODO(),
			deploymenName,
			types.StrategicMergePatchType,
			[]byte(patchPayload),
			metav1.PatchOptions{},
		)
		if err != nil {
			log.Fatal("Error occured in update func's patch call. Error is: ", err)
		}
	}
}
