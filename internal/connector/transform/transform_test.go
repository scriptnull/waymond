package transform

import (
	"testing"
)

func TestGoTemplate(t *testing.T) {
	tt := []struct {
		transformer  Transformer
		inputData    string
		expectsError bool
		expectedData string
	}{
		{
			transformer: &goTemplate{
				template: `{"asg_name": "{{ .queue }}","desired_count": {{ .scheduled_jobs_count }}}`,
			},
			inputData:    `{"queue":"aws-on-demand-amd64-ubuntu-xlarge","scheduled_jobs_count":1}`,
			expectsError: false,
			expectedData: `{"asg_name": "aws-on-demand-amd64-ubuntu-xlarge","desired_count": 1}`,
		},
	}

	for _, tc := range tt {
		outputBytes, err := tc.transformer.Transform([]byte(tc.inputData))
		if tc.expectsError && err == nil {
			t.Error("test case expects error, but didn't get back an error")
		}

		if !tc.expectsError && err != nil {
			t.Errorf("test case doesn't expect an error, but got back an error: %s", err)
		}

		outputData := string(outputBytes)
		if tc.expectedData != outputData {
			t.Errorf("expected data: %s, but got: %s", tc.expectedData, outputData)
		}
	}
}
