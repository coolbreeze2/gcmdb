package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetNoNamespaced(t *testing.T) {
	RootCmd.SetArgs([]string{"get", "deploytemplate"})
	assertOsExit(t, Execute, 1)
}

func TestGetResourceAll(t *testing.T) {
	cases := [][]string{
		{"apply", "-f", "../example/files"},
		{"get", "deploytemplate", "-A"},
	}

	ts := testServer()
	defer ts.Close()

	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
	}
}

func TestGetResource(t *testing.T) {

	resources := [][]string{
		{"secret", "test"},
		{"datacenter", "test"},
		{"zone", "test"},
		{"namespace", "test"},
		{"scm", "gitlab-test"},
		{"hostnode", "test"},
		{"helmrepository", "test"},
		{"containerregistry", "harbor-test"},
		{"configcenter", "apollo-test"},
		{"deployplatform", "test"},
		{"project", "go-devops"},
		{"app", "go-app"},
		{"deploytemplate", "docker-compose-test", "-n", "test"},
		{"resourcerange", "test", "-n", "test"},
		{"orchestration", "test"},
		{"appdeployment", "go-app", "-n", "test"},
	}

	cases := [][]string{
		{"apply", "-f", "../example/files"},
	}
	for _, r := range resources {
		ident := r[2:]
		c1 := append([]string{"get", r[0]}, ident...)
		cases = append(cases, c1)
		c2 := append([]string{"get", r[0], "-l", "x=y"}, ident...)
		cases = append(cases, c2)
		c3 := append([]string{"get", r[0], "-o", "yaml"}, ident...)
		cases = append(cases, c3)
		c4 := append([]string{"get", r[0], r[1], "-o", "yaml"}, ident...)
		cases = append(cases, c4)
		c5 := append([]string{"get", r[0], "-o", "json"}, ident...)
		cases = append(cases, c5)
		c6 := append([]string{"get", r[0], r[1], "-o", "json"}, ident...)
		cases = append(cases, c6)
	}

	ts := testServer()
	defer ts.Close()

	for i := range cases {
		RootCmd.SetArgs(cases[i])
		err := RootCmd.Execute()
		assert.NoError(t, err)
		if flag := RootCmd.PersistentFlags().Lookup("namespace"); flag != nil {
			flag.Value.Set("")
		}
	}
}

func TestHumanDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{d: -2 * time.Second, want: "<invalid>"},
		{d: -2*time.Second + 1, want: "0s"},
		{d: 0, want: "0s"},
		{d: time.Second - time.Millisecond, want: "0s"},
		{d: 2*time.Minute - time.Millisecond, want: "119s"},
		{d: 2 * time.Minute, want: "2m"},
		{d: 2*time.Minute + time.Second, want: "2m1s"},
		{d: 10*time.Minute - time.Millisecond, want: "9m59s"},
		{d: 10 * time.Minute, want: "10m"},
		{d: 10*time.Minute + time.Second, want: "10m"},
		{d: 3*time.Hour - time.Millisecond, want: "179m"},
		{d: 3 * time.Hour, want: "3h"},
		{d: 3*time.Hour + time.Minute, want: "3h1m"},
		{d: 8*time.Hour - time.Millisecond, want: "7h59m"},
		{d: 8 * time.Hour, want: "8h"},
		{d: 8*time.Hour + 59*time.Minute, want: "8h"},
		{d: 2*24*time.Hour - time.Millisecond, want: "47h"},
		{d: 2 * 24 * time.Hour, want: "2d"},
		{d: 2*24*time.Hour + time.Hour, want: "2d1h"},
		{d: 8*24*time.Hour - time.Millisecond, want: "7d23h"},
		{d: 8 * 24 * time.Hour, want: "8d"},
		{d: 8*24*time.Hour + 23*time.Hour, want: "8d"},
		{d: 2*365*24*time.Hour - time.Millisecond, want: "729d"},
		{d: 2 * 365 * 24 * time.Hour, want: "2y"},
		{d: 2*365*24*time.Hour + 23*time.Hour, want: "2y"},
		{d: 2*365*24*time.Hour + 23*time.Hour + 59*time.Minute, want: "2y"},
		{d: 2*365*24*time.Hour + 24*time.Hour - time.Millisecond, want: "2y"},
		{d: 2*365*24*time.Hour + 24*time.Hour, want: "2y1d"},
		{d: 3 * 365 * 24 * time.Hour, want: "3y"},
		{d: 4 * 365 * 24 * time.Hour, want: "4y"},
		{d: 5 * 365 * 24 * time.Hour, want: "5y"},
		{d: 6 * 365 * 24 * time.Hour, want: "6y"},
		{d: 7 * 365 * 24 * time.Hour, want: "7y"},
		{d: 8*365*24*time.Hour - time.Millisecond, want: "7y364d"},
		{d: 8 * 365 * 24 * time.Hour, want: "8y"},
		{d: 8*365*24*time.Hour + 364*24*time.Hour, want: "8y"},
		{d: 9 * 365 * 24 * time.Hour, want: "9y"},
	}
	for _, tt := range tests {
		t.Run(tt.d.String(), func(t *testing.T) {
			if got := HumanDuration(tt.d); got != tt.want {
				t.Errorf("HumanDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
