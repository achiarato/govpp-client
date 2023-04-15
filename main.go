package main

import (
	"strconv"
	"net"
	"fmt"
	"encoding/json"
        "io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/api"

//	interfaces "github.com/achiarato/GoVPP/vppbinapi/interface"
//	"github.com/achiarato/GoVPP/vppbinapi/interface_types"
	"github.com/achiarato/GoVPP/vppbinapi/ip_types"
	sr "github.com/achiarato/GoVPP/vppbinapi/sr"
	"github.com/achiarato/GoVPP/vppbinapi/sr_types"
//	"github.com/achiarato/GoVPP/vppbinapi/vpe"


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


func SrPolicyDump(ch api.Channel) error {
	fmt.Println("Dumping SR Policies installed on VPP")
	time.Sleep(1 * time.Second)

	n := 0
	reqCtx := ch.SendMultiRequest(&sr.SrPoliciesDump{})

	for {
		msg := &sr.SrPoliciesDetails{}
		stop, err := reqCtx.ReceiveReply(msg)
		if stop {
			break
		}
		if err != nil {
			return err
		}
		n++
		fmt.Printf(" - SR Policy #%d: \n", n)
		fmt.Printf("    BSID:      %+v\n", msg.Bsid)
		fmt.Printf("    IsSpray:   %+v\n", msg.IsSpray)
		fmt.Printf("    IsEncap:   %+v\n", msg.IsEncap)
		fmt.Printf("    Fib Table: %+v\n", msg.FibTable)
		fmt.Printf("    SID List:  %+v\n", msg.SidLists[0].Sids)
		//		fmt.Printf("   SID List:  %+v\n", Sids)
	}
	if n == 0 {
		fmt.Println("No Srv6 Policies configured")
	}
	return nil
}

func ToVppIP6Address(addr net.IP) ip_types.IP6Address {
	ip := [16]uint8{}
	copy(ip[:], addr)
	return ip
}

func ToVppAddress(addr net.IP) ip_types.Address {
	a := ip_types.Address{}
	if addr.To4() == nil {
		a.Af = ip_types.ADDRESS_IP6
		ip := [16]uint8{}
		copy(ip[:], addr)
		a.Un = ip_types.AddressUnionIP6(ip)
	} else {
		a.Af = ip_types.ADDRESS_IP4
		ip := [4]uint8{}
		copy(ip[:], addr.To4())
		a.Un = ip_types.AddressUnionIP4(ip)
	}
	return a
}

func ToVppPrefix(prefix *net.IPNet) ip_types.Prefix {
	len, _ := prefix.Mask.Size()
	r := ip_types.Prefix{
		Address: ToVppAddress(prefix.IP),
		Len:     uint8(len),
	}
	return r
}

func SrSteeringAddDel(ch api.Channel, Bsid ip_types.IP6Address, Traffic ip_types.Prefix) error {
	fmt.Println("Adding SR Steer policy")

	var traffic_type sr_types.SrSteer
	if Traffic.Address.Af == ip_types.ADDRESS_IP4 {
		traffic_type = 4
	} else {
		traffic_type = 6
	}

	request := &sr.SrSteeringAddDel{
		BsidAddr:    Bsid,
		TableID:     0,
		Prefix:      Traffic,
		SwIfIndex:   2,
		TrafficType: traffic_type,
	}

	response := &sr.SrSteeringAddDelReply{}
	err := ch.SendRequest(request).ReceiveReply(response)
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	fmt.Println("SRv6 Steer Policy added!")
	return nil
}

func SrPolicyAdd(ch api.Channel, Bsid ip_types.IP6Address, Isspray bool, Isencap bool, Fibtable int, Sids [16]ip_types.IP6Address, Sidslen int) error {

	fmt.Println("Adding SRv6 Policy")

	BSID := (Bsid)
	PolicyBsid := ip_types.IP6Address{}
	PolicyBsid = BSID
	FwdTable := Fibtable
	FibTable := uint32(FwdTable)

	request := &sr.SrPolicyAdd{
		BsidAddr: PolicyBsid,
		IsSpray:  Isspray,
		IsEncap:  Isencap,
		FibTable: FibTable,
		Sids: sr.Srv6SidList{
			NumSids: uint8(Sidslen),
			Weight:  1,
			Sids:    Sids,
		},
	}
	response := &sr.SrPolicyAddReply{}
	err := ch.SendRequest(request).ReceiveReply(response)
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	fmt.Println("SRv6 Policy added: ", Sids)
	return nil
}



func main() {
	// Connect to VPP
	conn, err := govpp.Connect("/var/run/vpp/vpp-api.sock")
	defer conn.Disconnect()
	if err != nil {
		fmt.Printf("Could not connect: %s\n", err)
		os.Exit(1)
	}

	// Open channel
	ch, err := conn.NewAPIChannel()
	defer ch.Close()
	if err != nil {
		fmt.Printf("Could not open API channel: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("GoVPP Client is ready to go!")
	time.Sleep(1*time.Second)
	fmt.Println("GoVPP Client will start polling for inputs coming from apps!")
	time.Sleep(3*time.Second)
	var oldapplist []string
	var vpp_bsid string
	vpp_bsid = "1::"
	app_counter := 1

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
					fmt.Println("Exporting data from the HTTP Response")
					time.Sleep(1*time.Second)
					fmt.Printf("Source Node: %s\n", response.Src)
					fmt.Printf("Destination Node: %s\n", response.Dst)
					fmt.Printf("USid: %s\n", response.USid)
					fmt.Printf("Query Type: %s\n", response.Query)
					time.Sleep(2*time.Second)
					fmt.Println("Configuring policies for VPP via GoVPP")
					time.Sleep(1*time.Second)
					new_bsid := vpp_bsid + strconv.Itoa(app_counter)
					policyBSID := ToVppIP6Address(net.ParseIP(new_bsid))
					app_counter += 1
					segments := [16]ip_types.IP6Address{}
					segments[0] = ToVppIP6Address(net.ParseIP(response.USid))
					err = SrPolicyAdd(ch, policyBSID, false, true, 0, segments, 1)
					if err != nil {
						fmt.Printf("Could not add SR Policy: %s\n", err)
						os.Exit(1)
					}
					time.Sleep(1*time.Second)
					addr, network, err := net.ParseCIDR("10.10.1.1/24")
					_ = addr
					traffic := ToVppPrefix(network)
					err = SrSteeringAddDel(ch, policyBSID, traffic)
					if err != nil {
						fmt.Printf("Could not add SR Steer Policy: %s\n", err)
						os.Exit(1)
					}
					time.Sleep(1*time.Second)
					err = SrPolicyDump(ch)
					if err != nil {
						fmt.Printf("Could not dump SR Policies: %s\n", err)
						os.Exit(1)
					}

					time.Sleep(4*time.Second)
				}
			} else {
				time.Sleep(3*time.Second)
				fmt.Printf("Keep polling. No new inputs to process. Inputs: %s\n", newapplist)
				//break
			}
		}

	}

}
