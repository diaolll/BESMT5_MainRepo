package service

import "testing"

func TestWithinCreditLimit(t *testing.T) {
	tests := []struct {
		name   string
		taken  int
		adding int
		want   bool
	}{
		{name: "masih di bawah batas", taken: 18, adding: 3, want: true},
		{name: "tepat di batas", taken: 21, adding: 3, want: true},
		{name: "melewati batas", taken: 22, adding: 3, want: false},
		{name: "kelas 6 SKS melewati batas", taken: 21, adding: 6, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinCreditLimit(tt.taken, tt.adding); got != tt.want {
				t.Fatalf("WithinCreditLimit(%d, %d) = %v, want %v",
					tt.taken, tt.adding, got, tt.want)
			}
		})
	}
}

func TestCreditsAfterSwap(t *testing.T) {
	tests := []struct {
		name    string
		taken   int
		leaving int
		joining int
		want    int
	}{
		{name: "SKS tetap", taken: 21, leaving: 3, joining: 3, want: 21},
		{name: "SKS naik", taken: 18, leaving: 3, joining: 6, want: 21},
		{name: "SKS turun", taken: 24, leaving: 6, joining: 3, want: 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreditsAfterSwap(tt.taken, tt.leaving, tt.joining); got != tt.want {
				t.Fatalf("CreditsAfterSwap(%d, %d, %d) = %d, want %d",
					tt.taken, tt.leaving, tt.joining, got, tt.want)
			}
		})
	}
}
