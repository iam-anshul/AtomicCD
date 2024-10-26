package pkg

import (
	"log"
)

func checkContainerConfig(containerConfig *[]map[string]interface{}, scope string) {
	for _, container := range *containerConfig {

		if container["containername"] == nil && scope == "namescoped" {
			log.Fatalf("'containerName' field not specified, incorrect configuration of target with targetName: %v", container["targetName"])
		}

		if container["containername"].(string) == "" && scope == "namescoped" {
			log.Fatalf("'containerName' field empty, incorrect configuration of target with targetName: %s", container["targetName"])
		}

		if container["containerimage"] == nil {
			log.Fatalf("'containerImage' field not specified, incorrect configuration of target with targetName: %v", container["targetName"])
		}

		if container["containerimage"].(string) == "" {
			log.Fatalf("'containerImage' field empty, incorrect configuration of target with targetName: %s", container["targetName"])
		}

		if container["containertag"] == nil {
			log.Fatalf("'containerTag' field not specified, incorrect configuration of target with targetName: %v", container["targetName"])
		}

		if container["containertag"].(string) == "" {
			log.Fatalf("'containerTag' field empty, incorrect configuration of target with targetName: %s", container["targetName"])
		}

		if scope == "nameScoped" {

			for key, _ := range container {
				if key != "containername" && key != "containerimage" && key != "containertag" && key != "targetName" {
					log.Fatalf("unknown field '%v', incorrect configuration of target with targetName: %v", key, container["targetName"])
				}
			}

		} else {

			for key, _ := range container {
				if key != "containerimage" && key != "containertag" && key != "targetName" {
					log.Fatalf("unknown field '%v', incorrect configuration of target with targetName: %v", key, container["targetName"])
				}
			}
		}

	}
}

func checkTargetConfigFile(targetConfigFile *[]map[string]interface{}) {
	for i := range len(*targetConfigFile) {

		if (*targetConfigFile)[i]["namespace"] == nil {
			log.Fatalf("'namespace' field not specified, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}

		if (*targetConfigFile)[i]["namespace"].(string) == "" {
			log.Fatalf("'Namespace' field empty, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}

		if (*targetConfigFile)[i]["targetname"] == nil {
			log.Fatalf("'targetName' field not specified, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}

		if (*targetConfigFile)[i]["targetname"] == "" {
			log.Fatalf("empty 'targetName' field, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}

		if (*targetConfigFile)[i]["scope"] == nil {
			log.Fatalf("'scope' field not specified, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}
		var scope string = (*targetConfigFile)[i]["scope"].(string)

		if scope == "namespaceScoped" {

			if (*targetConfigFile)[i]["type"] != nil {
				log.Fatalf("'type' field not allowed when scope is namespaceScoped, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
			}

			if (*targetConfigFile)[i]["name"] != nil {
				log.Fatalf("'name' field not allowed when scope is namespaceScoped, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
			}

			for key, _ := range (*targetConfigFile)[i] {
				if key != "targetname" && key != "containers" && key != "namespace" && key != "scope" {
					log.Fatalf("unknown field '%v', incorrect configuration of target with targetName: %v", key, (*targetConfigFile)[i]["targetname"])
				}
			}

		} else if scope == "nameScoped" {

			var deployType string = (*targetConfigFile)[i]["type"].(string)
			if deployType != "deployment" && deployType != "statefulset" {
				log.Fatalf("'type' filed can only be 'deployment' or 'statefulset' as its values, incorrect configuration of target with targetName: %s ", (*targetConfigFile)[i]["targetname"])
			}

			if (*targetConfigFile)[i]["name"] == nil {
				log.Fatalf("'name' field not specified, incorrect configuration of target with targetName: %s ", (*targetConfigFile)[i]["targetname"])
			}

			for key, _ := range (*targetConfigFile)[i] {
				if key != "targetname" && key != "containers" && key != "namespace" && key != "type" && key != "scope" && key != "name" {
					log.Fatalf("unknown field '%v', incorrect configuration of target with targetName: %v", key, (*targetConfigFile)[i]["targetname"])
				}
			}

		} else {
			log.Fatalf("'scope' field can only have 'namespaceScoped' or 'nameScoped' as its values, incorrect configuration of target with targetName: %s", (*targetConfigFile)[i]["targetname"])
		}

	}
}
