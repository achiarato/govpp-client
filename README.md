# govpp-client: GO Client for the end-host networking use case

End-Host VM interacting with Applications getting inputs to query a remote HTTP Server working as SRv6 SIDs Policy engine. 
GoVPP will use the received SRv6 SIDs to programme an SRv6 policy via VPP's APIs.

For more information about End-Host Networking:

Repository with GO code for the SRv6 SIDs Policy Engine:

DEMO: 

Interaction with Applications is simulated by GO Client polling a JSON file as following:
```
{
    "applications": [
        {
            "name": "Application Name",
            "requirement": "low latency",
            "src": "#IP Address Source/Topology Source Node",
            "dst": "#IP Address Destination/Topology Source Node"
        },
        {
            "name": "App1",
            "requirement": "low latency",
            "src": "2_0_0_0000.0000.0001",
            "dst": "2_0_0_0000.0000.0013"
        }
    ]
}
```
The JSON file will provide the Application Name, the Application Requirement (low latency, bandwidth guaranteed, ...), the source IP/Node and the destination IP/Node.

GO Client will wait until Applications provide inputs (JSON file filled). Once GO client will find inputs, it will prepare the HTTP Request based on Application requirements.

```
GoVPP Client is ready to go!
GoVPP Client will start polling for inputs coming from apps!
...
No inputs yet, still polling!
...
No inputs yet, still polling!
...
Hey! New application request is here!
Inputs coming from new app: App1. Starting the processing...
Application's inputs:
Name: App1 - Requirement: low latency - Source Node: 2_0_0_0000.0000.0001 - Destination Node: 2_0_0_0000.0000.0013

Preparing the HTTP request to query the SR-App...
HTTP request is ready: http://localhost:3333/shortestpath?src=2_0_0_0000.0000.0001&dst=2_0_0_0000.0000.0013
```
In the demo, the low latency input will be translated in the SHORTEST PATH requirement between Source and Destination node. (The SRv6 Policy Engine is locally installed in the demo but it could be remote)
The SRv6 Policy Engine will process the HTTP GET Request providing a SRv6 uSID in the HTTP Response's Body (for more information regarding SRv6 Policy Engine see above):

```
Sending HTTP request to the SR-App!
...
HTTP response received!
HTTP response's JSON Body: {"src":"2_0_0_0000.0000.0001","dst":"2_0_0_0000.0000.0013","uSid":"fc00::1:1:2:11:13","Query":"Shortest Path"}

Exporting data from the HTTP Response...
uSid received via HTTP response: fc00::1:1:2:11:13 for the Shortest Path query
```
The received uSID will be used to create new SRv6 policy programmed on VPP via GoVPP. 

```
Using data to configure SR policies for VPP via GoVPP...
Adding SRv6 Policy...
SRv6 Policy added:  [fc00::1:1:2:11:13 :: :: :: :: :: :: :: :: :: :: :: :: :: :: ::]

Adding SR Steer policy...
SR Steer Policy added!

Querying VPP to get list of configured policies to double check...
Dumping SR Policies...
 - SR Policy #1:
    BSID:      1::1
    IsSpray:   false
    IsEncap:   true
    Fib Table: 0
    SID List:  [fc00::1:1:2:11:13 :: :: :: :: :: :: :: :: :: :: :: :: :: :: ::]

Great! VPP config for App1 worked!
```
By adding a SR Steer Policy pointing the BSID 1::1 via GoVPP the networking configuration of the host based on application inputs is completed and traffic will be automatically steered accrodingly.

