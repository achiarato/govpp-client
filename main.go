package main

import (
	"fmt"
	"encoding/json"
        "io"
	"net/http"
	"os"
)

type ResponseBody struct {
     Src   string
     Dst   string
     USid  string
     Query string
}

const serverPort = 3333

func main() {

	requestURL := fmt.Sprintf("http://localhost:%d/shortestpath?src=2_0_0_0000.0000.0001&dst=2_0_0_0000.0000.0013", serverPort)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		fmt.Printf("Response failed\nStatus code: %d\nBody: %s\n", res.StatusCode, body)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("error reading body: %s", err)
		os.Exit(1)
	}
	fmt.Printf("JSON Body: %s\n", body)
	var response ResponseBody
	json.Unmarshal(body, &response)
	fmt.Printf("JSON.Unmarshal: %s\n", response)
	fmt.Printf("Source Node: %s\n", response.Src)
	fmt.Printf("Destination Node: %s\n", response.Dst)
	fmt.Printf("USid: %s\n", response.USid)
	fmt.Printf("Query Type: %s\n", response.Query)

}
