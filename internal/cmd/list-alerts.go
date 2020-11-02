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
	"strings"
	"sync"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

var listAlertsCmdSettings struct {
	id   string
	from string
	to   string
}

// currentShiftCmd represents the get command
var listAlertsCmd = &cobra.Command{
	Use:   "list-alerts",
	Args:  cobra.ExactArgs(0),
	Short: "Lists all alerts",
	Long:  `Lists all alerts in a specified time period`,
	RunE: func(cmd *cobra.Command, args []string) error {

		client, err := pd.CreatePagerDutyClient()
		if err != nil {
			return err
		}

		var user *pagerduty.User
		if listAlertsCmdSettings.id == "" {
			user, err = client.GetCurrentUser(pagerduty.GetCurrentUserOptions{})
			if err != nil {
				return err
			}
		} else {
			user, err = client.GetUser(listAlertsCmdSettings.id, pagerduty.GetUserOptions{})
			if err != nil {
				return err
			}
		}

		offset := 0
		limit := 100
		numPrintedIncidents := 0
		teamIDs := listTeamIDs(*user)
		moreIncidents := true
		for moreIncidents {

			listIncidentsOptions := pagerduty.ListIncidentsOptions{
				APIListObject: pagerduty.APIListObject{
					Limit:  uint(limit),
					Offset: uint(offset),
				},
				Since:   listAlertsCmdSettings.from,
				Until:   listAlertsCmdSettings.to,
				TeamIDs: teamIDs,
			}
			a, err := client.ListIncidents(listIncidentsOptions)
			incidents := a.Incidents
			if err != nil {
				return err
			}

			incidents, err = filterIncidentsByNameInLogEntries(incidents, user.Name, client)

			for _, incident := range incidents {
				bunt.Printf("\n%d. *%s*\n", numPrintedIncidents+1, incident.Title)

				if incident.Description != incident.Title {
					bunt.Printf("   *Description:* \n")
					bunt.Printf("   %s\n", incident.Description)
				}

				bunt.Printf("   *Link:* %s\n", strings.Replace(incident.Self, "api", "ibm", 1))

				start := mustParsePagerDutyRFC3339ishTime(incident.CreatedAt)

				end := mustParsePagerDutyRFC3339ishTime(incident.LastStatusChangeAt)

				bunt.Printf("   *Time:* %s - %s (%s)\n",
					start.Local().Format("2006-01-02 15:04:05"),
					end.Local().Format("2006-01-02 15:04:05"),
					end.Sub(start),
				)

				notes, err := client.ListIncidentNotes(incident.Id)
				if err != nil {
					return err
				}
				if len(notes) > 0 {
					bunt.Printf("   *Notes:*\n")
					for j := len(notes) - 1; j >= 0; j-- {
						bunt.Printf("      %d. %s (by _%s_ at _%s_)\n",
							len(notes)-j,
							notes[j].Content,
							lookUpNameByUserID(client, notes[j].User.ID),
							formatNoteTime(notes[j].CreatedAt),
						)
					}
				}

				numPrintedIncidents++
			}
			if err != nil {
				return err
			}

			if a.More {
				offset += limit
			} else {
				moreIncidents = false
			}
		}

		bunt.Println()

		return nil
	},
}

func lookUpNameByUserID(client *pagerduty.Client, id string) string {
	user, err := client.GetUser(id, pagerduty.GetUserOptions{})
	if err != nil {
		return "unknown"
	}
	return user.Name
}

func formatNoteTime(input string) string {
	time, err := time.Parse("2006-01-02T15:04:05-07:00", input)
	if err != nil {
		return input
	}

	return time.Local().Format("2006-01-02 15:04:05")
}

func filterIncidentsByNameInLogEntries(incidents []pagerduty.Incident, username string, client *pagerduty.Client) ([]pagerduty.Incident, error) {
	numIncidents := len(incidents)
	var filteredIncidents []pagerduty.Incident
	logEntries := make([]pagerduty.ListIncidentLogEntriesResponse, numIncidents, numIncidents)

	loadLogEntries := func(position int, wg *sync.WaitGroup) {
		logEntry, _ := client.ListIncidentLogEntries(incidents[position].Id, pagerduty.ListIncidentLogEntriesOptions{})
		logEntries[position] = *logEntry
		wg.Done()
	}
	var wg *sync.WaitGroup
	if numIncidents <= 50 { // if you sent 100 requests at once, you sometimes get an error. 50 requests at once is pretty safe.
		wg = new(sync.WaitGroup)
		wg.Add(numIncidents)

		for i := 0; i < numIncidents; i++ {
			go loadLogEntries(i, wg)
		}
	} else {
		wg = new(sync.WaitGroup)
		wg.Add(int(numIncidents / 2))

		for i := 0; i < int(numIncidents/2); i++ {
			go loadLogEntries(i, wg)
		}

		wg.Wait()
		wg.Add(int(numIncidents / 2))

		for i := int(numIncidents / 2); i < numIncidents; i++ {
			go loadLogEntries(i, wg)
		}
	}
	wg.Wait()

	for i, incident := range incidents {
		for _, logEntry := range logEntries[i].LogEntries {
			if strings.Contains(logEntry.CommonLogEntryField.Summary, username) {
				filteredIncidents = append(filteredIncidents, incident)
				break
			}
		}
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

func mustParsePagerDutyRFC3339ishTime(input string) time.Time {
	time, err := time.Parse("2006-01-02T15:04:05Z", input)
	if err != nil {
		panic(err)
	}
	return time
}

func init() {
	rootCmd.AddCommand(listAlertsCmd)

	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.id, "id", "", "use custom ID")
	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.from, "from", "", "set startpoint of custom time period")
	listAlertsCmd.Flags().StringVar(&listAlertsCmdSettings.to, "to", "", "set endpoint of custom time period")
}
