package main

import "testing"

func TestParseNodename(t *testing.T) {
	testName := "ns=1;s=[L2S2_TMCP]Lift_Station_Consume.Alarms[3]"
	expected := "L2S2_TMCP_Lift_Station_Consume_Alarms_3"
	result := nodeNameToMetricName(&testName)

	if result != expected {
		t.Errorf("GOT %s, WANTED %s", result, expected)
	}
}
