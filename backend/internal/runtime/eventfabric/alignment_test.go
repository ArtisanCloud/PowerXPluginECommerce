package eventfabric

import "testing"

func TestValidateConsistency(t *testing.T) {
	manifest := []TopicDecl{
		{Topic: "_topic.demo.a", Actions: []string{"publish", "subscribe"}},
		{Topic: "_topic.demo.b", Actions: []string{"publish"}},
	}
	execution := []TopicDecl{
		{Topic: "_topic.demo.a", Actions: []string{"subscribe", "publish"}},
		{Topic: "_topic.demo.b", Actions: []string{"publish"}},
	}
	if err := ValidateConsistency(manifest, execution); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateConsistencyMismatch(t *testing.T) {
	manifest := []TopicDecl{{Topic: "_topic.demo.a", Actions: []string{"publish"}}}
	execution := []TopicDecl{{Topic: "_topic.demo.a", Actions: []string{"subscribe"}}}
	if err := ValidateConsistency(manifest, execution); err == nil {
		t.Fatal("expected mismatch error")
	}
}
