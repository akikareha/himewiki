package format

import "testing"

func TestTrimForSummaryCrash(t *testing.T) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but function returned normally")
		}
	}()

	text := "This is a test."
	_ = TrimForSummary(text, -1)
}

func TestTrimForSummaryCrash2(t *testing.T) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but function returned normally")
		}
	}()

	text := "This is a test."
	_ = TrimForSummary(text, -10)
}

func TestTrimForSummary(t *testing.T) {
	var text, got, want string

	text = "This is a test."
	got = TrimForSummary(text, 0)
	want = ""
	if got != want {
		t.Errorf("TrimForSummary(text, 0) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 0)
	want = ""
	if got != want {
		t.Errorf("TrimForSummary(text, 0) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 1)
	want = "."
	if got != want {
		t.Errorf("TrimForSummary(text, 1) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 1)
	want = "."
	if got != want {
		t.Errorf("TrimForSummary(text, 1) = %s; want %s", got, want)
	}

	text = "A"
	got = TrimForSummary(text, 1)
	want = "A"
	if got != want {
		t.Errorf("TrimForSummary(text, 1) = %s; want %s", got, want)
	}

	text = "あ"
	got = TrimForSummary(text, 1)
	want = "あ"
	if got != want {
		t.Errorf("TrimForSummary(text, 1) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 2)
	want = ".."
	if got != want {
		t.Errorf("TrimForSummary(text, 2) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 2)
	want = ".."
	if got != want {
		t.Errorf("TrimForSummary(text, 2) = %s; want %s", got, want)
	}

	text = "AB"
	got = TrimForSummary(text, 2)
	want = "AB"
	if got != want {
		t.Errorf("TrimForSummary(text, 2) = %s; want %s", got, want)
	}

	text = "あい"
	got = TrimForSummary(text, 2)
	want = "あい"
	if got != want {
		t.Errorf("TrimForSummary(text, 2) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 3)
	want = "T.."
	if got != want {
		t.Errorf("TrimForSummary(text, 3) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 3)
	want = "こ.."
	if got != want {
		t.Errorf("TrimForSummary(text, 3) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 10)
	want = "This is .."
	if got != want {
		t.Errorf("TrimForSummary(text, 10) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 5)
	want = "これは.."
	if got != want {
		t.Errorf("TrimForSummary(text, 10) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 14)
	want = "This is a te.."
	if got != want {
		t.Errorf("TrimForSummary(text, 14) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 8)
	want = "これはテスト.."
	if got != want {
		t.Errorf("TrimForSummary(text, 14) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 15)
	want = "This is a test."
	if got != want {
		t.Errorf("TrimForSummary(text, 15) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 9)
	want = "これはテストです。"
	if got != want {
		t.Errorf("TrimForSummary(text, 15) = %s; want %s", got, want)
	}

	text = "This is a test."
	got = TrimForSummary(text, 20)
	want = "This is a test."
	if got != want {
		t.Errorf("TrimForSummary(text, 20) = %s; want %s", got, want)
	}

	text = "これはテストです。"
	got = TrimForSummary(text, 15)
	want = "これはテストです。"
	if got != want {
		t.Errorf("TrimForSummary(text, 20) = %s; want %s", got, want)
	}
}
