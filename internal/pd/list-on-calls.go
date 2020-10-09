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
	"path/filepath"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/mitchellh/go-homedir"
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

// GetPagerDutyOnCalls returns all on-calls for the current user
func GetPagerDutyOnCalls(client *pagerduty.Client) (map[TimeRange]map[string]pagerduty.EscalationPolicy, error) {
	user, err := client.GetCurrentUser(pagerduty.GetCurrentUserOptions{})
	if err != nil {
		return nil, err
	}

	listOptions := pagerduty.ListOnCallOptions{
		APIListObject: pagerduty.APIListObject{Limit: 100},
		UserIDs:       []string{user.ID},
		Earliest:      true,
	}

	list, err := client.ListOnCalls(listOptions)
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

func loadConfig() (*Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filepath.Join(home, ".pd.yml"))
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func parsePagerDutyTime(input string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", input)
}
