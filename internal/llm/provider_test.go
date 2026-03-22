package llm

import (
	"context"
	"testing"
)

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleSystem, "system"},
		{RoleUser, "user"},
		{RoleAssistant, "assistant"},
		{RoleTool, "tool"},
	}

	for _, tt := range tests {
		if string(tt.role) != tt.want {
			t.Errorf("Role = %q, want %q", tt.role, tt.want)
		}
	}
}

// mockProvider verifies Provider can be implemented and exercised.
type mockProvider struct{}

func (m *mockProvider) Complete(_ context.Context, _ []Message, _ []Tool) (*Response, error) {
	return &Response{Content: "mock"}, nil
}

func (m *mockProvider) Stream(_ context.Context, _ []Message, _ []Tool) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestProviderComplete(t *testing.T) {
	var p Provider = &mockProvider{}

	resp, err := p.Complete(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "mock" {
		t.Fatalf("Content = %q, want %q", resp.Content, "mock")
	}
}

func TestProviderStreamSendsDoneChunk(t *testing.T) {
	var p Provider = &mockProvider{}

	ch, err := p.Stream(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	var gotDone bool
	for chunk := range ch {
		if chunk.Done {
			gotDone = true
		}
	}
	if !gotDone {
		t.Fatal("stream must send a Done chunk before closing")
	}
}
