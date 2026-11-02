package script

import (
	"testing"
	"time"
)

func TestOptionsDefaultsAndValidation(t *testing.T) {
	o, err := (Options{}).Normalize()
	if err != nil || o.Timeout != 5*time.Minute || o.MaxRows != 1000 || o.MaxBytes != 16<<20 {
		t.Fatalf("%+v %v", o, err)
	}
	for _, o := range []Options{{Timeout: -1}, {MaxRows: -1}, {MaxBytes: -1}} {
		if _, err = o.Normalize(); err == nil {
			t.Fatalf("accepted %+v", o)
		}
	}
}
