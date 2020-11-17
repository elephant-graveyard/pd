// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"errors"
	"strings"
	"sync"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/wrap"
	"github.com/homeport/pd/internal/pd"
)

func getRelevantIncidents(userID string, from string, to string) ([]pagerduty.Incident, string, error) {
	client, err := pd.CreatePagerDutyClient()
	if err != nil {
		return nil, "", err
	}

	var user *pagerduty.User
	if userID == "" {
		user, err = client.GetCurrentUser(pagerduty.GetCurrentUserOptions{})
		if err != nil {
			return nil, user.Name, wrap.Error(err, "it seems like the authtoken is not set correctly or outdated. Please update the authtoken in the .pd.yml file. If you don't know how to create your authtoken, this might help:\n https://support.pagerduty.com/docs/generating-api-keys#generating-a-personal-rest-api-key\n")
		}
	} else {
		user, err = client.GetUser(userID, pagerduty.GetUserOptions{})
		if err != nil {
			return nil, user.Name, wrap.Error(err, "it seems like the authtoken is not set correctly/outdated or the user-ID is invalid. Please update the authtoken in the .pd.yml file or use another user-ID. If you don't know how to create your authtoken, this might help:\n https://support.pagerduty.com/docs/generating-api-keys#generating-a-personal-rest-api-key\n")
		}
	}

	incidents := []pagerduty.Incident{}
	offset := 0
	limit := 100
	teamIDs := listTeamIDs(*user)
	moreIncidents := len(teamIDs) > 0
	if !moreIncidents {
		return nil, user.Name, errors.New("this PagerDuty-account is not part of any teams. To use this function, the PagerDuty-account must be part of at least one team")
	}
	for moreIncidents {
		listIncidentsOptions := pagerduty.ListIncidentsOptions{
			APIListObject: pagerduty.APIListObject{
				Limit:  uint(limit),
				Offset: uint(offset),
			},
			Since:   from,
			Until:   to,
			TeamIDs: teamIDs,
		}
		a, err := client.ListIncidents(listIncidentsOptions)
		if err != nil {
			return incidents, user.Name, err
		}
		incs, err := filterIncidentsByNameInLogEntries(a.Incidents, user.Name, client)
		if err != nil {
			return incidents, user.Name, err
		}
		incidents = append(incidents, incs...)

		if a.More {
			offset += limit
		} else {
			moreIncidents = false
		}
	}

	return incidents, user.Name, nil
}

func filterIncidentsByNameInLogEntries(incidents []pagerduty.Incident, username string, client *pagerduty.Client) ([]pagerduty.Incident, error) {
	const parallel = 10

	type in struct {
		index    int
		incident pagerduty.Incident
	}

	var (
		errors     = []error{}
		tasks      = make(chan in, len(incidents))
		logEntries = make([]pagerduty.ListIncidentLogEntriesResponse, len(incidents))
	)

	// Fill the task channel with work to be done
	for idx, incident := range incidents {
		tasks <- in{idx, incident}
	}

	// Start a worker group to process work tasks
	var wg sync.WaitGroup
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				resp, err := client.ListIncidentLogEntries(task.incident.Id, pagerduty.ListIncidentLogEntriesOptions{})
				if err != nil {
					errors = append(errors, err)
				}
				logEntries[task.index] = *resp
			}
		}()
	}
	close(tasks)
	wg.Wait()

	var filteredIncidents []pagerduty.Incident
	for i, incident := range incidents {
		for _, logEntry := range logEntries[i].LogEntries {
			if strings.Contains(logEntry.CommonLogEntryField.Summary, username) {
				filteredIncidents = append(filteredIncidents, incident)
				break
			}
		}
	}
	if len(errors) > 0 {
		return filteredIncidents, wrap.Errors(errors, "failed to filter incidents by name")
	}
	return filteredIncidents, nil

}

func listTeamIDs(user pagerduty.User) []string {
	result := make([]string, len(user.Teams))
	for i, team := range user.Teams {
		result[i] = team.ID
	}
	return result
}
