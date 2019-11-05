/*
 Copyright 2019 The Tekton Authors
 Copyright 2019 Orange

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

const (
	// Environment variable containing GitHub secret token
	envSecret = "GITHUB_SECRET_TOKEN"
)

type Repository struct {
	Name     *string `json:"name"`
	URL      *string `json:"url"`
	Revision *string `json:"revision"`
	Branch   *string `json:"branch"`
	FullName *string `json:"fullName"`
}

type Result struct {
	EventType   string     `json:"eventType"`
	EventAction string     `json:"eventAction"`
	Repository  Repository `json:"repository"`
}

func main() {
	secretToken := os.Getenv(envSecret)
	if secretToken == "" {
		log.Fatalf("No secret token given")
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//TODO: We should probably send over the EL eventID as a X-Tekton-Event-Id header as well
		payload, err := github.ValidatePayload(request, []byte(secretToken))
		id := github.DeliveryID(request)
		if err != nil {
			log.Printf("Error handling Github Event with delivery ID %s : %q", id, err)
			http.Error(writer, fmt.Sprint(err), http.StatusBadRequest)
			return
		}
		log.Printf("Handling Github Event with delivery ID: %s; Payload: %s", id, payload)
		event, err := github.ParseWebHook(github.WebHookType(request), payload)
		if err != nil {
			log.Printf("Failed to parse Github event ID: %s. Error: %q", id, err)
			http.Error(writer, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		result := Result{}
		eventName := request.Header.Get("X-GitHub-Event")
		switch event := event.(type) {
		case *github.PullRequestEvent:
			result.EventType = eventName
			result.EventAction = *event.Action
			result.Repository = Repository{
				Name:     event.Repo.Name,
				FullName: event.Repo.FullName,
				URL:      event.PullRequest.Head.Repo.CloneURL,
				Revision: event.PullRequest.Head.SHA,
				Branch:   event.PullRequest.Head.Ref,
			}
		case *github.PushEvent:
			branch := strings.Split(*event.Ref, "/")[2]
			result.EventType = eventName
			result.Repository = Repository{
				Name:     event.Repo.Name,
				FullName: event.Repo.FullName,
				URL:      event.Repo.CloneURL,
				Revision: event.HeadCommit.ID,
				Branch:   &branch,
			}
			if pb := request.Header.Get("X-Push-Branches-Only"); pb != "" {
				branches := strings.Split(pb, ",")
				match := false
				for _, b := range branches {
					if b == branch {
						match = true
					}
				}
				if !match {
					log.Printf("Aborting since push is not on master branch but: %s", branch)
					http.Error(writer, fmt.Sprint("Disallowed push event on this branch"), http.StatusLocked)
					return
				}
			}
		default:
			log.Printf("Event %s not supported", eventName)
			http.Error(writer, fmt.Sprintf("Event %s not supported", eventName), http.StatusBadRequest)
			return
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Printf("Failed to marshal JSON result: %s", err)
			http.Error(writer, fmt.Sprint(err), http.StatusInternalServerError)
		}
		n, err := writer.Write(jsonResult)
		if err != nil {
			log.Printf("Failed to write response for Github event ID: %s. Bytes writted: %d. Error: %q", id, n, err)
		}
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8080), nil))
}
