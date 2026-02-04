package common

import "testing"

type TestCase struct {
	Name string
	Run  func(t *testing.T)
}

func RunTestCases(t *testing.T, tests []TestCase) {
	t.Helper()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Run(t)
		})
	}
}
