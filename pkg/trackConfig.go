package pkg

import (
	"log"

	"github.com/spf13/viper"
)

func GetTrackConfig() *map[string]interface{} {
	TrackConfig := viper.New()

	TrackConfig.SetConfigFile("/atomicCD/config/trackConfig.yaml")
	if err := TrackConfig.ReadInConfig(); err != nil {
		log.Fatalf("Error reading TrackConfig.yaml file: %v", err)
	}

	var tConfig map[string]interface{}

	err := TrackConfig.UnmarshalKey("trackConfig", &tConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling TrackConfig: %v", err)
	}

	checkTrackConfig(&tConfig)

	return &tConfig
}

func checkTrackConfig(tConfig *map[string]interface{}) {

	if (*tConfig)["repourl"] == nil {
		log.Fatalf("'repoURL' field not specified, incorrect tracK config configuration.")
	}

	if (*tConfig)["repourl"].(string) == "" {
		log.Fatalf("'repoURL' field empty, incorrect tracK config configuration.")
	}

	if (*tConfig)["path"] == nil {
		log.Fatalf("'path' field not specified, incorrect tracK config configuration.")
	}

	if (*tConfig)["path"].(string) == "" {
		log.Fatalf("'path' field empty, incorrect tracK config configuration.")
	}

	for key, _ := range *tConfig {
		if key != "repourl" && key != "path" && key != "branch" && key != "webhooksecret" && key != "token" {
			log.Fatalf("Unknown field in trac config %s", key)
		}
	}
}
