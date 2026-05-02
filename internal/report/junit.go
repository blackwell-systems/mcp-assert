// junit.go generates JUnit XML reports compatible with CI systems (GitHub
// Actions, Jenkins, GitLab CI). When the suite was run with multiple trials
// (--trials > 1), pass@k and pass^k reliability properties are included
// in the <testsuite> element.

package report

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/blackwell-systems/mcp-assert/internal/assertion"
)

// JUnit XML types following the standard schema.

type junitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Skipped    int              `xml:"skipped,attr"`
	Time       float64          `xml:"time,attr"`
	Properties *junitProperties `xml:"properties,omitempty"`
	Cases      []junitTestCase  `xml:"testcase"`
}

type junitProperties struct {
	Properties []junitProperty `xml:"property"`
}

type junitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Skipped   *junitSkipped `xml:"skipped,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr,omitempty"`
}

// WriteJUnit writes JUnit XML to the given path.
func WriteJUnit(results []assertion.Result, path string) error {
	var totalTime assertion.DurationMS
	var failures, skipped int
	var cases []junitTestCase

	for _, r := range results {
		totalTime += r.Duration

		tc := junitTestCase{
			Name:      r.Name,
			Classname: "mcp-assert",
			Time:      r.Duration.Seconds(),
		}
		if r.Language != "" {
			tc.Classname = r.Language
		}

		switch r.Status {
		case assertion.StatusFail:
			failures++
			tc.Failure = &junitFailure{
				Message: r.Detail,
				Text:    r.Detail,
			}
		case assertion.StatusSkip:
			skipped++
			tc.Skipped = &junitSkipped{}
		}

		cases = append(cases, tc)
	}

	suite := junitTestSuite{
		Name:     "mcp-assert",
		Tests:    len(results),
		Failures: failures,
		Skipped:  skipped,
		Time:     totalTime.Seconds(),
		Cases:    cases,
	}

	// Embed pass@k / pass^k as <property> elements when the suite was run
	// with multiple trials. CI systems can parse these for flaky-test reporting.
	if hasMultipleTrials(results) {
		stats := ComputeReliability(results)
		capable, reliable := 0, 0
		for _, s := range stats {
			if s.PassAt {
				capable++
			}
			if s.PassUp {
				reliable++
			}
		}
		total := len(stats)
		suite.Properties = &junitProperties{
			Properties: []junitProperty{
				{Name: "pass_at_k", Value: fmt.Sprintf("%d/%d", capable, total)},
				{Name: "pass_up_k", Value: fmt.Sprintf("%d/%d", reliable, total)},
			},
		}
	}

	suites := junitTestSuites{
		Suites: []junitTestSuite{suite},
	}

	data, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JUnit XML: %w", err)
	}

	output := []byte(xml.Header)
	output = append(output, data...)
	output = append(output, '\n')

	return os.WriteFile(path, output, 0644)
}

// hasMultipleTrials returns true if any result has Trial > 1.
func hasMultipleTrials(results []assertion.Result) bool {
	for _, r := range results {
		if r.Trial > 1 {
			return true
		}
	}
	return false
}
