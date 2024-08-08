package ganalytics

import "testing"

func TestDecodeData(t *testing.T) {
	paylaod, err := DecodedData("eyJ0cmFja2luZyI6eyJldmVudCI6Ii8iLCJjYXRlZ29yeSI6IlBhZ2Ugdmlld3MiLCJpZGVudGl0eSI6IiIsInJlZmVycmVyIjoiIiwidXNlckFnZW50IjoiTW96aWxsYS81LjAgKFgxMTsgVWJ1bnR1OyBMaW51eCB4ODZfNjQ7IHJ2OjEyOC4wKSBHZWNrby8yMDEwMDEwMSBGaXJlZm94LzEyOC4wIiwiaXNUb3VjaERldmljZSI6ZmFsc2UsInR5cGUiOiJwYWdlIn0sInNpdGVfaWQiOiJnb25hbHl0aWNzIn0=");
	if err != nil {
		t.Fatal(err)
	}else if paylaod.SiteID != "gonalytics" {
		t.Errorf("expected [gonalytics] got [%s]", paylaod.SiteID)
	}
}