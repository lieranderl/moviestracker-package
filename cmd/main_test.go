package main

import "testing"

func TestBuildTrackerURLNormalizesKinozalQueryAndScheme(t *testing.T) {
	template := "http://kinozal.tv/browse.php?s=%s (1080p|2160p)&g=3&c=0&v=0&d=%s&w=0&t=0&f=0"

	got, err := buildTrackerURL(template, "Bad Boys", "2024")
	if err != nil {
		t.Fatalf("buildTrackerURL returned error: %v", err)
	}

	want := "https://kinozal.tv/browse.php?c=0&d=2024&f=0&g=3&s=Bad+Boys+%281080p%7C2160p%29&t=0&v=0&w=0"
	if got != want {
		t.Fatalf("unexpected URL\nwant: %s\ngot:  %s", want, got)
	}
}

func TestBuildTrackerURLNormalizesRutorPath(t *testing.T) {
	template := "http://rutor.is/search/0/0/300/0/%s %s (2160p|1080p)"

	got, err := buildTrackerURL(template, "Bad Boys", "2024")
	if err != nil {
		t.Fatalf("buildTrackerURL returned error: %v", err)
	}

	want := "http://rutor.is/search/0/0/300/0/Bad%20Boys%202024%20%282160p%7C1080p%29"
	if got != want {
		t.Fatalf("unexpected URL\nwant: %s\ngot:  %s", want, got)
	}
}
