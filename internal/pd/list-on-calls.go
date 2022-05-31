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

package pd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/gonvenience/wrap"
	"gopkg.in/yaml.v3"
)

// TimeRange specifies a time range
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// CreatePagerDutyClient creates a new PagerDuty client based on the access
// token stored in the ~/.pd.yml file
func CreatePagerDutyClient() (*pagerduty.Client, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return pagerduty.NewClient(config.Authtoken), nil
}

// GetPagerDutyOnCalls returns all currently active on-calls for the user
func GetPagerDutyOnCalls(client *pagerduty.Client, user *pagerduty.User) (map[TimeRange]map[string]pagerduty.EscalationPolicy, error) {
	list, err := GetAllOnCalls(client, user, "", "")
	if err != nil {
		return nil, err
	}

	var oncalls = map[TimeRange]map[string]pagerduty.EscalationPolicy{}
	for i := range list.OnCalls {
		oncall := list.OnCalls[i]

		start, err := parsePagerDutyTime(oncall.Start)
		if err != nil {
			return nil, err
		}

		end, err := parsePagerDutyTime(oncall.End)
		if err != nil {
			return nil, err
		}

		timeRange := TimeRange{start, end}

		if _, found := oncalls[timeRange]; !found {
			oncalls[timeRange] = map[string]pagerduty.EscalationPolicy{}
		}

		oncalls[timeRange][oncall.EscalationPolicy.APIObject.ID] = oncall.EscalationPolicy
	}

	return oncalls, nil
}

// GetAllOnCalls returns all on calls for a specified user in a specified time range
// If time range is not specified, only currently active on-calls will be returned
func GetAllOnCalls(client *pagerduty.Client, user *pagerduty.User, start string, end string) (*pagerduty.ListOnCallsResponse, error) {
	return client.ListOnCalls(
		pagerduty.ListOnCallOptions{
			Limit:    100,
			UserIDs:  []string{user.ID},
			Since:    start,
			Until:    end,
			Earliest: true,
		})
}

// GetTemplate returns the requested template
func GetTemplate(templateName string) (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", err
	}
	return config.Templates[templateName], err
}

func loadConfig() (*Config, error) {

	data, err := getDataFromYAMLFile()
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, wrap.Error(err, "it seems like the content of the .pd.yml file could not be interpreted. Please follow these instructions to set up the file correctly: https://github.com/homeport/pd/blob/main/README.md")
	}

	return &config, nil
}

func getDataFromYAMLFile() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filepath.Join(home, ".pd.yml"))
	if err != nil {
		return nil, wrap.Error(err, "it seems like the .pd.yml is not created or could not be read. Please follow these instructions to set up the file correctly: https://github.com/homeport/pd/blob/main/README.md")
	}
	return data, nil
}

func parsePagerDutyTime(input string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", input)
}
