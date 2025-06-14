// main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
	"gopkg.in/yaml.v3"
)

const (
	prometheusURL = "http://localhost:9090"
	query 	      = "avg(rate(container_cpu_usage_seconds_total%5B1m%5D))%20by%20(pod)"
	threshold     = 0.00003
	repoPath      = "charts/nginx"
	valuesFile    = "values.yaml"
	targetPod     = "nginx-auto-tuned" // substring match
)

func main() {
	for {
		fmt.Println("Starting CPU usage check...")
		cpuUsage := fetchCPUUsage()
		fmt.Printf("Fetched CPU usage for target pod: %.8f\n", cpuUsage)
		if cpuUsage > threshold {
			fmt.Println("Threshold breached. Updating values.yaml...")
			updateValuesYaml()
			commitAndPush()
		} else {
			fmt.Println("CPU usage below threshold. No update needed.")
		}
		time.Sleep(60 * time.Second)
	}
}

func fetchCPUUsage() float64 {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, query)
	fmt.Println("Querying:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching from Prometheus:", err)
		return 0
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Prometheus response:", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if res, ok := data["result"].([]interface{}); ok {
			for _, entry := range res {
				m := entry.(map[string]interface{})
				metric := m["metric"].(map[string]interface{})
				podName := metric["pod"].(string)
				if bytes.Contains([]byte(podName), []byte(targetPod)) {
					valStr := m["value"].([]interface{})[1].(string)
					var val float64
					fmt.Printf("Matched pod: %s\n", podName)
					fmt.Sscanf(valStr, "%f", &val)
					return val
				}
			}
		}
	}
	fmt.Println("No matching pod found with substring:", targetPod)
	return 0
}

func updateValuesYaml() {
	fmt.Println("Modifying values.yaml...")
	file := fmt.Sprintf("%s/%s", repoPath, valuesFile)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Failed to read values.yaml:", err)
		return
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		fmt.Println("Failed to parse YAML:", err)
		return
	}

	current := 1
	if v, ok := values["replicaCount"].(int); ok {
		current = v
	} else if v, ok := values["replicaCount"].(float64); ok {
		current = int(v)
	}

	values["replicaCount"] = current + 1

	newData, err := yaml.Marshal(values)
	if err != nil {
		fmt.Println("Failed to serialize YAML:", err)
		return
	}

	if err := ioutil.WriteFile(file, newData, 0644); err != nil {
		fmt.Println("Failed to write values.yaml:", err)
		return
	}

	fmt.Printf("replicaCount incremented from %d to %d\n", current, current+1)
}

func commitAndPush() {
	fmt.Println("Committing and pushing to Git...")
	cmds := [][]string{
		{"git", "-C", repoPath, "add", "."},
		{"git", "-C", repoPath, "commit", "-m", "auto: tuned replicas due to high CPU"},
		{"git", "-C", repoPath, "push"},
	}
	for _, cmdArgs := range cmds {
		fmt.Printf("Running: %v\n", cmdArgs)
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error: %s\nOutput: %s\n", err, output)
		} else {
			fmt.Printf("Output: %s\n", output)
		}
	}
	fmt.Println("Git push completed.")
}

