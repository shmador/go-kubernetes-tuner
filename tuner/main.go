package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

const (
	prometheusURL = "http://localhost:9090"
	query         = "avg(rate(container_cpu_usage_seconds_total[1m])) by (pod)"
	threshold     = 0.75
	repoPath      = "../charts/nginx"
	valuesFile    = "values.yaml"
)

func main() {
	for {
		cpuUsage := fetchCPUUsage()
		if cpuUsage > threshold {
			fmt.Println("Threshold breached. Updating values.yaml...")
			updateValuesYaml()
			commitAndPush()
		}
		time.Sleep(60 * time.Second)
	}
}

func fetchCPUUsage() float64 {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, query))
	if err != nil {
		fmt.Println("Error fetching from Prometheus:", err)
		return 0
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	// Minimal parsing: grab first value
	if data, ok := result["data"].(map[string]interface{}); ok {
		if res, ok := data["result"].([]interface{}); ok && len(res) > 0 {
			valStr := res[0].(map[string]interface{})["value"].([]interface{})[1].(string)
			var val float64
			fmt.Sscanf(valStr, "%f", &val)
			return val
		}
	}
	return 0
}

func updateValuesYaml() {
	file := fmt.Sprintf("%s/%s", repoPath, valuesFile)
	input, _ := ioutil.ReadFile(file)
	output := bytes.Replace(input, []byte("replicaCount: 3"), []byte("replicaCount: 5"), -1)
	ioutil.WriteFile(file, output, 0644)
}

func commitAndPush() {
	cmds := [][]string{
		{"git", "-C", repoPath, "add", "."},
		{"git", "-C", repoPath, "commit", "-m", "auto: tuned replicas due to high CPU"},
		{"git", "-C", repoPath, "push"},
	}
	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error: %s\nOutput: %s\n", err, output)
		}
	}
}

