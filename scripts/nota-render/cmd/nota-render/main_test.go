package main

import "testing"

func TestParseArgsAllowsFlagsAfterJob(t *testing.T) {
	args, err := parseArgs([]string{"paper", "--jobs", "4", "--force", "--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Jobname != "paper" || args.Jobs != "4" || !args.Force || !args.Verbose {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestParseArgsAllowsJobsEquals(t *testing.T) {
	args, err := parseArgs([]string{"--jobs=auto", "paper"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Jobname != "paper" || args.Jobs != "auto" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
