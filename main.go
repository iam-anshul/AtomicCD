package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iam-anshul/atomicCD/pkg"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset
var TrackConfig *map[string]interface{} = pkg.GetTrackConfig()

func createClientSet() (*kubernetes.Clientset, error) {
	serviceAccountToken, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Fatal(err)
	}

	kubeconfig := &rest.Config{
		Host:        "https://kubernetes.default.svc",
		BearerToken: string(serviceAccountToken),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	return kubernetes.NewForConfig(kubeconfig)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("recieved webhook request")

	if (*TrackConfig)["webhooksecret"] != nil {

		fmt.Println("Authenticating webhook...")
		secret := (*TrackConfig)["webhooksecret"]
		signature := r.Header.Get("X-Hub-Signature-256")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		mac := hmac.New(sha256.New, []byte(secret.(string)))
		mac.Write(body)

		expectedMac := "sha256=" + fmt.Sprintf("%x", mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMac)) {
			http.Error(w, "Invalid signature", http.StatusForbidden)
			fmt.Println("Webhook authentication failed!")
			return
		}
		fmt.Println("Webhook authentication successful!")
	}

	fmt.Println("Recieved webhook trigger. Syncing")

	time.Sleep(3 * time.Second)

	pkg.Execute(clientset)
}

func main() {
	var err error
	clientset, err = createClientSet()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		http.HandleFunc("/webhook", webhook)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	for {
		fmt.Println("running")
		pkg.Execute(clientset)
		time.Sleep(180 * time.Second)
	}
}
