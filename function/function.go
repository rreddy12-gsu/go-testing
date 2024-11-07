// Package cxwh contains an example Dialogflow CX webhook
package cxwh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"slices"
	"strings"

	"cloud.google.com/go/storage"
)

type item struct {
	Name       string   `json:"name"`
	Price      float32  `json:"price"`
	Categories []string `json:"categories"`
}

type store struct {
	Name  string `json:"name"`
	Items []item `json:"items"`
}

type fulfillmentInfo struct {
	Tag string `json:"tag"`
}

type sessionInfo struct {
	Session    string                 `json:"session"`
	Parameters map[string]interface{} `json:"parameters"`
}

type text struct {
	Text []string `json:"text"`
}

type responseMessage struct {
	Text text `json:"text"`
}

type fulfillmentResponse struct {
	Messages []responseMessage `json:"messages"`
}

// webhookRequest is used to unmarshal a WebhookRequest JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by the Dialogflow protocol buffers:
// https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/dialogflow/cx/v3#WebhookRequest
type webhookRequest struct {
	FulfillmentInfo fulfillmentInfo `json:"fulfillmentInfo"`
	SessionInfo     sessionInfo     `json:"sessionInfo"`
}

// webhookResponse is used to marshal a WebhookResponse JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by the Dialogflow protocol buffers:
// https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/dialogflow/cx/v3#WebhookResponse
type webhookResponse struct {
	FulfillmentResponse fulfillmentResponse `json:"fulfillmentResponse"`
	SessionInfo         sessionInfo         `json:"sessionInfo"`
}

// confirm handles webhook calls using the "confirm" tag.
func confirm(request webhookRequest) (webhookResponse, error) {
	// Create a text message that utilizes the "size" and "color"
	// parameters provided by the end-user.
	// This text message is used in the response below.
	t := fmt.Sprintf("You can pick up your order for a %s %s shirt in 5 days.",
		request.SessionInfo.Parameters["size"],
		request.SessionInfo.Parameters["color"])

	// Create session parameters that are populated in the response.
	// The "cancel-period" parameter is referenced by the agent.
	// This example hard codes the value 2, but a real system
	// might look up this value in a database.
	p := map[string]interface{}{"cancel-period": "2"}

	// Build and return the response.
	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{t},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil
}

func findItem(request webhookRequest) (webhookResponse, error) {
	t := ""

	stores := getStores("stores-test", "stores")

	item := request.SessionInfo.Parameters["item"]

	storesWith := []store{}
	for _, store := range stores {
		for _, i := range store.Items {
			if i.Name == item {
				storesWith = append(storesWith, store)
			}
		}
	}

	storeNames := []string{}

	for _, store := range storesWith {
		storeNames = append(storeNames, store.Name)
	}

	t += "This item is sold at: "
	t += strings.Join(storeNames, ", ")
	t += fmt.Sprint(storeNames)

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{t},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: map[string]interface{}{"cancel-period": "2"},
		},
	}
	return response, nil
}

func listStores(request webhookRequest) (webhookResponse, error) {

	t := ""

	stores := getStores("stores-test", "stores")

	storeNames := []string{}
	for _, store := range stores {
		if !slices.Contains(storeNames, store.Name) {
			storeNames = append(storeNames, store.Name)
		}
	}

	t += "This item is sold at: "
	t += strings.Join(storeNames, ", ")

	p := map[string]interface{}{"stores": stores}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{t},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}

	return response, nil
}

func cheapest(request webhookRequest) (webhookResponse, error) {
	t := ""

	stores := getStores("stores-test", "stores")

	item := request.SessionInfo.Parameters["item"]

	cheapestStore := store{}
	var price float32 = math.MaxFloat32
	for _, store := range stores {
		for _, i := range store.Items {
			if i.Name == item {
				if i.Price <= price {
					price = i.Price
					cheapestStore = store
				}
			}
		}
	}

	t += fmt.Sprintf("it is sold at %s for %.2f", cheapestStore.Name, price)

	p := map[string]interface{}{"cheapest": cheapestStore}

	response := webhookResponse{
		FulfillmentResponse: fulfillmentResponse{
			Messages: []responseMessage{
				{
					Text: text{
						Text: []string{t},
					},
				},
			},
		},
		SessionInfo: sessionInfo{
			Parameters: p,
		},
	}
	return response, nil
}

func getStores(bucket string, object string) []store {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println("could not create client")
	}

	bkt := client.Bucket(bucket)

	obj := bkt.Object(object)

	r, err := obj.NewReader(ctx)
	if err != nil {
		fmt.Println("could not create reader")
	}

	data, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("could not read")
	}
	var stores []store
	json.Unmarshal(data, &stores)
	fmt.Println(stores)
	r.Close()

	return stores
}

// handleError handles internal errors.
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

// HandleWebhookRequest handles WebhookRequest and sends the WebhookResponse.
func HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	var request webhookRequest
	var response webhookResponse
	var err error

	// Read input JSON
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Request: %+v", request)

	// Get the tag from the request, and call the corresponding
	// function that handles that tag.
	// This example only has one possible tag,
	// but most agents would have many.
	switch tag := request.FulfillmentInfo.Tag; tag {
	case "confirm":
		response, err = confirm(request)
	case "list":
		response, err = listStores(request)
	case "find":
		response, err = findItem(request)
	case "cheapest":
		response, err = cheapest(request)
	default:
		err = fmt.Errorf("Unknown tag: %s", tag)
	}
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Response: %+v", response)

	// Send response
	if err = json.NewEncoder(w).Encode(&response); err != nil {
		handleError(w, err)
		return
	}
}
