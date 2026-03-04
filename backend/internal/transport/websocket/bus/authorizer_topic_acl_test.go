package bus

import "testing"

func TestAuthorizerRejectsUnregisteredTopic(t *testing.T) {
	DefaultTopicRegistry = NewTopicRegistry()
	DefaultACLRegistry = NewACLRegistry()

	a := NewDefaultAuthorizer()
	client := &Client{TenantUUID: "tenant-a"}
	if err := a.Authorize(nil, client, "_topic.demo.updated"); err == nil {
		t.Fatal("expected topic rejection")
	}
}

func TestAuthorizerRequiresSubscribeGrant(t *testing.T) {
	DefaultTopicRegistry = NewTopicRegistry()
	DefaultACLRegistry = NewACLRegistry()
	DefaultTopicRegistry.Register([]string{"_topic.demo.updated"})
	DefaultACLRegistry.Grant("tenant-a", "_topic.demo.updated", []string{"publish"})

	a := NewDefaultAuthorizer()
	client := &Client{TenantUUID: "tenant-a"}
	if err := a.Authorize(nil, client, "_topic.demo.updated"); err == nil {
		t.Fatal("expected subscribe permission rejection")
	}

	DefaultACLRegistry.Grant("tenant-a", "_topic.demo.updated", []string{"subscribe"})
	if err := a.Authorize(nil, client, "_topic.demo.updated"); err != nil {
		t.Fatalf("expected allow after subscribe grant, got %v", err)
	}
}
