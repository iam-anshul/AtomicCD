package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var trackConfig *map[string]interface{} = GetTrackConfig()
var targetConfig *[]map[string]interface{} = updateTargetConfig()
var wg = sync.WaitGroup{}

func updateTargetConfig() *[]map[string]interface{} {

	repoURL := (*trackConfig)["repourl"].(string)
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		log.Fatalf("Error parsing repoURL: %v", err)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		log.Fatalf("Invalid repoURL: %s", repoURL)
	}

	username := pathParts[0]
	repo := pathParts[1]

	var branch string
	if (*trackConfig)["branch"] != nil {
		branch = (*trackConfig)["branch"].(string)
	} else {
		branch = "main" // Default to "main"
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", username, repo, branch, (*trackConfig)["path"])

	var body []byte
	if (*trackConfig)["token"] != nil {
		token := (*trackConfig)["token"].(string)

		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			log.Fatalf("Error occured in fetching targetConfig.yaml from repository %v", err)
		}

		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Cache-Control", "no-cache")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error occured in client request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Fatalf("Request to fetch target config failed with status code: %d", resp.StatusCode)
		}

		var ioError error
		body, ioError = io.ReadAll(resp.Body)
		if ioError != nil {
			log.Fatalf("Error occured in reading trackConfig: %v", err)
		}

	} else {

		resp, _ := http.Get(rawURL)
		if err != nil {
			log.Fatalf("Error occured in client request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Fatalf("Request to fetch target config failed with status code: %d", resp.StatusCode)
		}

		var ioError error
		body, ioError = io.ReadAll(resp.Body)
		if ioError != nil {
			log.Fatalf("Error occured in reading trackConfig: %v", err)
		}

	}

	targetConfigfile := viper.New()
	targetConfigfile.SetConfigType("yaml")
	k := targetConfigfile.ReadConfig(bytes.NewBuffer(body))
	if k != nil {
		log.Fatalf("error loading configuration from body: %v", err)
	}

	// Get the target configurations as a slice of interface{}
	var targets []map[string]interface{}
	err = targetConfigfile.UnmarshalKey("targetConfig", &targets)
	if err != nil {
		log.Fatalf("Error unmarshalling targetConfig: %v", err)
	}

	return &targets
}

func scanConfig() {
	targetConfig = updateTargetConfig()
}

func getContainers(i int) *[]map[string]interface{} {
	containers, err := (*targetConfig)[i]["containers"].([]interface{})
	if !err {
		log.Fatalf("There is no 'containers' field in targetName: %v", (*targetConfig)[i]["targetname"])
	}

	containerConfigList := make([]map[string]interface{}, 0)
	var targetName string = (*targetConfig)[i]["targetname"].(string)

	for container := range containers {

		containerConfig := containers[container].(map[string]interface{})
		containerConfig["targetName"] = targetName
		containerConfigList = append(containerConfigList, containerConfig)
	}
	checkContainerConfig(&containerConfigList)
	return &containerConfigList
}

func compareAndDeploy(i int, cluster *kubernetes.Clientset) {

	var namespace string = (*targetConfig)[i]["namespace"].(string)
	var deployType string

	if value, ok := (*targetConfig)[i]["type"].(string); ok {
		deployType = value
	}

	containerConfigList := getContainers(i)

	if (*targetConfig)[i]["scope"] == "namespacescoped" {

		deployments, err := cluster.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, deploy := range deployments.Items {
			var outOfSync bool = false
			var podInfo []map[string]string
			for _, currentContainer := range deploy.Spec.Template.Spec.Containers {

				currentImageConfig := strings.Split(currentContainer.Image, ":")
				currentImage := currentImageConfig[0]
				currentTag := currentImageConfig[1]

				for _, configContainer := range *containerConfigList {

					if configContainer["containerimage"] == currentImage && configContainer["containertag"] != currentTag {
						fmt.Printf("Out of sync Deployment: %s with Container Name %s\n", deploy.Name, currentContainer.Name)
						outOfSync = true

						containerInfo := map[string]string{
							"containerName":  currentContainer.Name,
							"containerImage": currentImage,
							"containerTag":   configContainer["containertag"].(string),
						}

						podInfo = append(podInfo, containerInfo)

					}
				}
			}

			if outOfSync {
				updateInNamespace(cluster, &podInfo, deploy.Name, namespace, "deployment")
				fmt.Println("Sync Successfull!")
			}
		}

		statefulsets, err := cluster.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, statefulset := range statefulsets.Items {
			var outOfSync bool = false
			var podInfo []map[string]string
			for _, currentContainer := range statefulset.Spec.Template.Spec.Containers {

				currentImageConfig := strings.Split(currentContainer.Image, ":")
				currentImage := currentImageConfig[0]
				currentTag := currentImageConfig[1]

				for _, configContainer := range *containerConfigList {

					if configContainer["containerimage"] == currentImage && configContainer["containertag"] != currentTag {
						fmt.Printf("Out of sync Deployment: %s with Container Name %s\n", statefulset.Name, currentContainer.Name)
						outOfSync = true

						containerInfo := map[string]string{
							"containerName":  currentContainer.Name,
							"containerImage": currentImage,
							"containerTag":   configContainer["containertag"].(string),
						}

						podInfo = append(podInfo, containerInfo)

					}
				}
			}

			if outOfSync {
				updateInNamespace(cluster, &podInfo, statefulset.Name, namespace, "statefulset")
				fmt.Println("Sync Successfull!")
			}
		}

	} else if deployType == "deployment" {
		deployments, err := cluster.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})

		if err != nil {
			log.Fatal(err)
		}

		for _, deploy := range deployments.Items {
			if deploy.Name == (*targetConfig)[i]["name"] {
				var outOfSync bool = false
				for _, currentContainer := range deploy.Spec.Template.Spec.Containers {

					currentImageConfig := strings.Split(currentContainer.Image, ":")
					currentImage := currentImageConfig[0]
					currentTag := currentImageConfig[1]

					for _, configContainer := range *containerConfigList {

						if configContainer["containername"] == currentContainer.Name && (configContainer["containertag"] != currentTag || configContainer["containerimage"] != currentImage) {

							fmt.Printf("Out of sync Deployment: %s with Container Name %s\n", deploy.Name, currentContainer.Name)
							update(cluster, containerConfigList, namespace, deploy.Name, deployType)
							fmt.Println("Sync Successfull!")
							outOfSync = true
							break
						}
					}

					if outOfSync {
						break
					}

				}
			}
		}
	} else if deployType == "statefulset" {
		statefulsets, err := cluster.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, statefulset := range statefulsets.Items {

			if statefulset.Name == (*targetConfig)[i]["name"] {
				var outOfSync bool = false
				for _, currentContainer := range statefulset.Spec.Template.Spec.Containers {

					currentImageConfig := strings.Split(currentContainer.Image, ":")
					currentImage := currentImageConfig[0]
					currentTag := currentImageConfig[1]

					for _, configContainer := range *containerConfigList {

						if configContainer["containername"] == currentContainer.Name && (configContainer["containertag"] != currentTag || configContainer["containerimage"] != currentImage) {

							fmt.Printf("Out of sync Deployment: %s with Container Name %s\n", statefulset.Name, currentContainer.Name)
							update(cluster, containerConfigList, namespace, statefulset.Name, deployType)
							fmt.Println("Sync Successfull!")
							outOfSync = true
							break

						}
					}

					if outOfSync {
						break
					}

				}
			}
		}

	}
	wg.Done()
}

func Execute(cluster *kubernetes.Clientset) {
	scanConfig()
	checkTargetConfigFile(targetConfig)

	for i := range len((*targetConfig)) {
		wg.Add(1)
		go compareAndDeploy(i, cluster)
	}
	wg.Wait()
}
