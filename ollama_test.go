package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	key := "2781b57b34a2415381d3277a76a7dae0.Q7aS0L8zgH34VH1COBqzBOgN"
	
	// Try native /api/chat no auth
	testNative(key)
	// Try Bearer auth
	testBearer(key)
	// Try x-api-key
	testXApiKey(key)
}

func testNative(key string) {
	url := "https://ollama.com/api/chat"
	body, _ := json.Marshal(map[string]interface{}{
		"model": "glm-5.1:cloud",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
		"stream": false,
	})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil { fmt.Printf("native err: %v\n", err); return }
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("native %d: %s\n", resp.StatusCode, string(data))
}

func testBearer(key string) {
	url := "https://ollama.com/api/chat"
	body, _ := json.Marshal(map[string]interface{}{
		"model": "glm-5.1:cloud",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
		"stream": false,
	})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { fmt.Printf("bearer err: %v\n", err); return }
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("bearer %d: %s\n", resp.StatusCode, string(data))
}

func testXApiKey(key string) {
	url := "https://ollama.com/api/chat"
	body, _ := json.Marshal(map[string]interface{}{
		"model": "glm-5.1:cloud",
		"messages": []map[string]string{{"role": "user", "content": "hi"}},
		"stream": false,
	})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { fmt.Printf("x-api-key err: %v\n", err); return }
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("x-api-key %d: %s\n", resp.StatusCode, string(data))
}
