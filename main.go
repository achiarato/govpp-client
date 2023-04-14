package main

import (
	"bufio"
	"strings"
	"strconv"
	"fmt"
	"encoding/json"
        "io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Applications struct {
     Apps []App `json:"applications"`
}

type App struct {
     Name string   `json:"name"`
     Req string   `json:"requirement"`
     Appsrc string `json:"src"`
     Appdst string `json:"dst"`
}

type ResponseBody struct {
     Src   string
     Dst   string
     USid  string
     Query string
}

const serverPort = 3333
var apps Applications


func main() {
	fmt.Println("GoVPP Client is ready to go!")
	time.Sleep(3*time.Second)
	fmt.Println("GoVPP Client will start polling for inputs coming from apps!")
	time.Sleep(3*time.Second)
	var oldapplist []string

	for {
		time.Sleep(3*time.Second)
		fmt.Println("...")
		jsonAppDB, err := os.Open("app.json")
		if err != nil {
			fmt.Printf("Error reading the appDB json file: %s", err)
			os.Exit(1)
		}
		defer jsonAppDB.Close()
		//_, err = jsonAppDB.Seek(0,0)
		AppDB, _ := ioutil.ReadAll(jsonAppDB)
		json.Unmarshal(AppDB, &apps)

		var newapplist []string
		if len(apps.Apps)==0 {
			time.Sleep(3*time.Second)
			fmt.Println("No inputs yet, still polling!")
			continue
		} else {//if len(apps.Apps)!=0 {
			for i:=0; i<len(apps.Apps); i++ {
				newapplist = append(newapplist, apps.Apps[i].Name)
			}
			difference := make([]string, 0)
			for _, v1 := range newapplist {
				p := false
				for _, v2 := range oldapplist {
					if v1 == v2 {
						p = true
						break
					}
				}
				if !p {
					difference = append(difference, v1)
				}
			}
			if len(difference) != 0 {
				time.Sleep(1*time.Second)
				fmt.Println("Hey! New application request is here!")
				time.Sleep(1*time.Second)
				fmt.Println("Processing new request!")
				time.Sleep(2*time.Second)
				//fmt.Printf("oldapplist: %s\n", oldapplist)
				oldapplist = newapplist
				//fmt.Printf("newapplist: %s\n", newapplist)
				//fmt.Printf("difference: %s\n", difference)
				var Todoapp Applications
				for t:=0; t<len(apps.Apps); t++ {
					for f:=0; f<len(difference); f++ {
						if apps.Apps[t].Name == difference[f] {
							Todoapp.Apps = append(Todoapp.Apps, apps.Apps[t])
						}
					}
				}

				for _, c := range Todoapp.Apps {
					fmt.Printf("Starting the processing for %s\n", c.Name)
					time.Sleep(1*time.Second)
					fmt.Printf("App Name: %s\n", c.Name)
					fmt.Printf("App Requ: %s\n", c.Req)
					fmt.Printf("App Src: %s\n", c.Appsrc)
					fmt.Printf("App Dst: %s\n", c.Appdst)
					time.Sleep(1*time.Second)
				
					url := "http://localhost:" + strconv.Itoa(serverPort) + "/"
					if c.Req == "low latency" {
						url += "shortestpath?"
					} else {
						fmt.Printf("error selecting the query type: no query type available")
						os.Exit(1)
					}
					url += "src=" + c.Appsrc + "&dst=" + c.Appdst
					fmt.Printf("Sending HTTP request %s\n", url)

					//inputs just to pause the automatic process
					fmt.Println("BREAKPOINT. Press RETURN to continue the process")
					reader := bufio.NewReader(os.Stdin)
                			input, err := reader.ReadString('\n')
			                if err != nil {
                        			fmt.Println("An error occured while reading input. Please try again", err)
                        			continue
                			}
                			input = strings.TrimSuffix(input, "\n")
					//finisce input

					res, err := http.Get(url)
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
					time.Sleep(1*time.Second)
					fmt.Println("HTTP Response received!")
					time.Sleep(1*time.Second)
					fmt.Printf("JSON Body: %s\n", body)
					var response ResponseBody
					json.Unmarshal(body, &response)
					time.Sleep(1*time.Second)
					fmt.Printf("JSON.Unmarshal: %s\n", response)
					time.Sleep(1*time.Second)
					fmt.Println("Exporting data from the HTTP Response")
					time.Sleep(1*time.Second)
					fmt.Printf("Source Node: %s\n", response.Src)
					fmt.Printf("Destination Node: %s\n", response.Dst)
					fmt.Printf("USid: %s\n", response.USid)
					fmt.Printf("Query Type: %s\n", response.Query)

					//inputs just to pause the automatic process
					fmt.Println("BREAKPOINT. Press RETURN to continue the process")
					reader = bufio.NewReader(os.Stdin)
                			input, err = reader.ReadString('\n')
			                if err != nil {
                        			fmt.Println("An error occured while reading input. Please try again", err)
                        			continue
                			}
                			input = strings.TrimSuffix(input, "\n")
					//finisce input

					time.Sleep(4*time.Second)

				}
			} else {
				time.Sleep(3*time.Second)
				fmt.Printf("Keep polling. No new inputs to process. Inputs: %s\n", newapplist)
				//break
			}
		}

	}
	os.Exit(1)
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
