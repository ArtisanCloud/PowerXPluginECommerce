package bus

import "testing"

func TestACLRegistryGrantAndAllowed(t *testing.T) {
	reg := NewACLRegistry()
	reg.Grant("tenant-a", "_topic.demo.updated", []string{"publish"})

	if !reg.Allowed("tenant-a", "_topic.demo.updated", "publish") {
		t.Fatal("expected publish to be allowed")
	}
	if reg.Allowed("tenant-a", "_topic.demo.updated", "subscribe") {
		t.Fatal("did not expect subscribe to be allowed")
	}
}

func TestACLRegistryDefaultActions(t *testing.T) {
	reg := NewACLRegistry()
	reg.Grant("tenant-a", "_topic.demo.updated", nil)

	if !reg.Allowed("tenant-a", "_topic.demo.updated", "publish") {
		t.Fatal("expected default publish action")
	}
	if !reg.Allowed("tenant-a", "_topic.demo.updated", "subscribe") {
		t.Fatal("expected default subscribe action")
	}
}
