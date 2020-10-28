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
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/bunt"
	"github.com/homeport/pd/internal/pd"
	"github.com/spf13/cobra"
)

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

		id, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}
		var user *pagerduty.User
		if id == "" {
			user, err = client.GetCurrentUser(pagerduty.GetCurrentUserOptions{})
			if err != nil {
				return err
			}
			id = user.ID
		} else {
			user, err = client.GetUser(id, pagerduty.GetUserOptions{})
		}

		start, err := cmd.Flags().GetString("from")
		if err != nil {
			return err
		}
		end, err := cmd.Flags().GetString("to")
		if err != nil {
			return err
		}
		startTime, err := time.Parse("2006-01-02T15:04:05Z", start)
		if err != nil {
			return err
		}
		endTime, err := time.Parse("2006-01-02T15:04:05Z", end)
		if err != nil {
			return err
		}

		useLogEntryMethod := true
		var ocs *pagerduty.ListOnCallsResponse
		if endTime.Sub(startTime).Hours() > 16 {
			useLogEntryMethod = false
			ocs, err = pd.GetAllOnCalls(client, user, start, end)
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
				Since:   start,
				Until:   end,
				TeamIDs: teamIDs,
			}
			a, err := client.ListIncidents(listIncidentsOptions)
			incidents := a.Incidents
			if err != nil {
				return err
			}

			if useLogEntryMethod {
				incidents, err = filterIncidentsByNameInLogEntries(incidents, user.Name, client)
			} else {
				incidents, err = filterIncidentsByTeamNameAndTime(incidents, ocs.OnCalls)
			}

			for _, incident := range incidents {
				bunt.Printf("\n%d. *%s*\n", numPrintedIncidents+1, incident.Title)

				if incident.Description != incident.Title {
					bunt.Printf("   *Description:* \n")
					bunt.Printf("   %s\n", incident.Description)
				}

				bunt.Printf("   *Link:* %s\n", incident.Self)

				start, err := time.Parse("2006-01-02T15:04:05Z", incident.CreatedAt)
				if err != nil {
					return err
				}

				end, err := time.Parse("2006-01-02T15:04:05Z", incident.LastStatusChangeAt)
				if err != nil {
					return err
				}

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

func filterIncidentsByTeamNameAndTime(incidents []pagerduty.Incident, oncalls []pagerduty.OnCall) ([]pagerduty.Incident, error) {
	var filteredIncidents []pagerduty.Incident
	for _, incident := range incidents {
		for _, team := range incident.Teams {
			for _, oncall := range oncalls {
				isInTimeSpan, err := inTimeSpan(oncall.Start, oncall.End, incident.CreatedAt)
				if err != nil {
					return filteredIncidents, err
				}
				if isInTimeSpan && strings.Contains(oncall.Schedule.Summary, team.Summary) {
					filteredIncidents = append(filteredIncidents, incident)
					break
				}
			}
		}
	}
	return filteredIncidents, nil
}

func inTimeSpan(s, e, check string) (bool, error) {
	start, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		return false, err
	}
	end, err := time.Parse("2006-01-02T15:04:05Z", e)
	if err != nil {
		return false, err
	}
	time, err := time.Parse("2006-01-02T15:04:05Z", check)
	if err != nil {
		return false, err
	}
	if start.Before(end) {
		return !time.Before(start) && !time.After(end), nil
	}
	if start.Equal(end) {
		return time.Equal(start), nil
	}
	return !start.After(time) || !end.Before(time), nil
}

func filterIncidentsByNameInLogEntries(incidents []pagerduty.Incident, username string, client *pagerduty.Client) ([]pagerduty.Incident, error) {
	var filteredIncidents []pagerduty.Incident
	for _, incident := range incidents {
		incLogEntries, err := client.ListIncidentLogEntries(incident.Id, pagerduty.ListIncidentLogEntriesOptions{})
		if err != nil {
			return nil, err
		}
		for _, logEntry := range incLogEntries.LogEntries {
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

func init() {
	rootCmd.AddCommand(listAlertsCmd)
	listAlertsCmd.Flags().String("id", "", "use custom ID")
	listAlertsCmd.Flags().String("from", "", "set startpoint of custom time period")
	listAlertsCmd.Flags().String("to", "", "set endpoint of custom time period")
}
