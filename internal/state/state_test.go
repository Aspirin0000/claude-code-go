package state

import (
	"testing"
)

func TestAppStateAddAndGetMessages(t *testing.T) {
	s := NewAppState()
	s.AddMessage(Message{Type: "user", Content: "hello"})
	s.AddMessage(Message{Type: "assistant", Content: "hi"})

	msgs := s.GetMessages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("unexpected first message content: %s", msgs[0].Content)
	}
	if msgs[1].Content != "hi" {
		t.Errorf("unexpected second message content: %s", msgs[1].Content)
	}
}

func TestAppStateClearMessages(t *testing.T) {
	s := NewAppState()
	s.AddMessage(Message{Type: "user", Content: "hello"})
	s.ClearMessages()

	msgs := s.GetMessages()
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}

func TestAppStateSetMessages(t *testing.T) {
	s := NewAppState()
	s.SetMessages([]Message{
		{Type: "system", Content: "sys"},
	})

	msgs := s.GetMessages()
	if len(msgs) != 1 || msgs[0].Type != "system" {
		t.Errorf("unexpected messages: %+v", msgs)
	}
}

func TestAppStateSessionAndCWD(t *testing.T) {
	s := NewAppState()
	s.SetSessionID("sess-123")
	s.SetCWD("/tmp/test")

	if s.SessionID != "sess-123" {
		t.Errorf("unexpected session id: %s", s.SessionID)
	}
	if s.CWD != "/tmp/test" {
		t.Errorf("unexpected cwd: %s", s.CWD)
	}
}

func TestAppStateIncrementTurn(t *testing.T) {
	s := NewAppState()
	s.IncrementTurn()
	s.IncrementTurn()

	if s.TurnCount != 2 {
		t.Errorf("expected turn count 2, got %d", s.TurnCount)
	}
}

func TestAppStateConcurrency(t *testing.T) {
	s := NewAppState()
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			s.AddMessage(Message{Type: "user", Content: "msg"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			s.GetMessages()
		}
		done <- true
	}()

	<-done
	<-done

	msgs := s.GetMessages()
	if len(msgs) != 100 {
		t.Errorf("expected 100 messages, got %d", len(msgs))
	}
}

func TestGlobalStateExists(t *testing.T) {
	if GlobalState == nil {
		t.Error("GlobalState should be initialized")
	}
}
