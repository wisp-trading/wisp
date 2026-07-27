package services

import "testing"

func TestFetchAvailableStrategies_StarterOnly(t *testing.T) {
	list, err := FetchAvailableStrategies()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 starter template, got %d", len(list))
	}
	if list[0].SDKExample != "starter" {
		t.Fatalf("got %q", list[0].SDKExample)
	}
}

func TestFormatProjectCreatedMsg(t *testing.T) {
	msg := FormatProjectCreatedMsg("demo", "starter")
	if msg == "" || !contains(msg, "demo") || !contains(msg, "Settings") {
		t.Fatalf("bad msg: %q", msg)
	}
}
