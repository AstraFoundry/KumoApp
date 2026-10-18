package application

import (
	"context"
	"errors"

	"golang.org/x/sys/windows"
)

// TUNPermission says whether the core can create a TUN interface.
type TUNPermission struct {
	Granted bool   `json:"granted"`
	Method  string `json:"method,omitempty"`
}

func tunPermission(string) TUNPermission {
	if windows.GetCurrentProcessToken().IsElevated() {
		return TUNPermission{Granted: true, Method: "administrator"}
	}
	return TUNPermission{}
}

func grantTUN(context.Context, string) error {
	return errors.New("on Windows, TUN needs Kumo to run as administrator")
}
