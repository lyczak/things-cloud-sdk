package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	thingscloud "github.com/nicolai86/things-cloud-sdk"
)

func stringVal(s string) *string {
	return &s
}

type ThingsTask struct {
	Title        string                `json:"title"`
	Note         string                `json:"note"`
	CreationDate thingscloud.Timestamp `json:"creationDate"`
	DeadlineDate thingscloud.Timestamp `json:"deadlineDate"`
}

type StatusResponse struct {
	Status   string   `json:"status"`
	Messages []string `json:"messages,omitempty"`
}

func main() {
	c := thingscloud.New(thingscloud.APIEndpoint, os.Getenv("THINGS_USERNAME"), os.Getenv("THINGS_PASSWORD"))

	if os.Getenv("THINGS_CONFIRMATION_CODE") != "" {
		if err := c.Accounts.Confirm(os.Getenv("THINGS_CONFIRMATION_CODE")); err != nil {
			log.Fatalf("Confirmation failed: %v", err.Error())
		}
		log.Printf("confirmation succeeded")
		return
	}

	_, err := c.Verify()
	if err != nil {
		log.Fatalf("Login failed: %q\nPlease check your credentials.", err.Error())
	}
	fmt.Printf("User: %s\n", c.EMail)

	history, err := c.OwnHistory()
	if err != nil {
		log.Fatalf("Failed to lookup own history key: %q\n", err.Error())
	}
	fmt.Printf("Own History Key: %s\n", history.ID)

	// BEGIN HTTP SERVER

	var status StatusResponse
	status.Status = "ok"
	status.Messages = make([]string, 0)

	http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(status)
	})

	http.HandleFunc("/tasks", func(writer http.ResponseWriter, request *http.Request) {
		var item ThingsTask
		json.NewDecoder(request.Body).Decode(&item)

		history.Sync()

		pending := thingscloud.TaskStatusPending
		taskUUID := uuid.New().String()

		var response StatusResponse

		log.Printf("Creating task %s\n", taskUUID)
		if err := history.Write(thingscloud.TaskActionItem{
			Item: thingscloud.Item{
				Kind:   thingscloud.ItemKindTask,
				Action: thingscloud.ItemActionCreated,
				UUID:   taskUUID,
			},
			P: thingscloud.TaskActionItemPayload{
				Title: stringVal(item.Title),
				Note:  stringVal(item.Note),
				//Schedule:     &anytime,
				Status:       &pending,
				CreationDate: &item.CreationDate,
				DeadlineDate: &item.DeadlineDate,
			},
		}); err == nil {
			writer.WriteHeader(http.StatusOK)
			response.Status = "ok"
		} else {
			errStr := fmt.Sprintf("Task creation failed: %q\n", err.Error())

			writer.WriteHeader(http.StatusInternalServerError)
			response.Status = "error"
			status.Messages = []string{errStr}

			log.Printf(errStr)

			status.Status = "error"
			status.Messages = append(status.Messages, errStr)
		}

		json.NewEncoder(writer).Encode(response)
	})

	http.ListenAndServe(":3000", nil)

	//state := memory.NewState()

	//anytime := thingscloud.TaskScheduleAnytime
	//yes := thingscloud.Boolean(true)

}
