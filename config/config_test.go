package config

import "testing"

func TestParseUserRolesEnv(t *testing.T) {
	roles := parseUserRolesEnv(" Tai.NguyenTanTai21042004@hcmut.edu.vn = admin ; analyst@hcmut.edu.vn:analyst")

	if got := roles["tai.nguyentantai21042004@hcmut.edu.vn"]; got != "ADMIN" {
		t.Fatalf("admin role = %q, want ADMIN", got)
	}
	if got := roles["analyst@hcmut.edu.vn"]; got != "ANALYST" {
		t.Fatalf("analyst role = %q, want ANALYST", got)
	}
}

func TestParseUserRolesEnvJSON(t *testing.T) {
	roles := parseUserRolesEnv(`{"admin@hcmut.edu.vn":"admin"}`)

	if got := roles["admin@hcmut.edu.vn"]; got != "ADMIN" {
		t.Fatalf("json role = %q, want ADMIN", got)
	}
}
