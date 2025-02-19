package imapwire

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/emersion/go-imap/v2"
)

const max = ^uint32(0)

func TestParseNumRange(t *testing.T) {
	tests := []struct {
		in  string
		out imap.NumRange
		ok  bool
	}{
		// Invalid number
		{"", imap.NumRange{}, false},
		{" ", imap.NumRange{}, false},
		{"A", imap.NumRange{}, false},
		{"0", imap.NumRange{}, false},
		{" 1", imap.NumRange{}, false},
		{"1 ", imap.NumRange{}, false},
		{"*1", imap.NumRange{}, false},
		{"1*", imap.NumRange{}, false},
		{"-1", imap.NumRange{}, false},
		{"01", imap.NumRange{}, false},
		{"0x1", imap.NumRange{}, false},
		{"1 2", imap.NumRange{}, false},
		{"1,2", imap.NumRange{}, false},
		{"1.2", imap.NumRange{}, false},
		{"4294967296", imap.NumRange{}, false},

		// Valid number
		{"*", imap.NumRange{0, 0}, true},
		{"1", imap.NumRange{1, 1}, true},
		{"42", imap.NumRange{42, 42}, true},
		{"1000", imap.NumRange{1000, 1000}, true},
		{"4294967295", imap.NumRange{max, max}, true},

		// Invalid range
		{":", imap.NumRange{}, false},
		{"*:", imap.NumRange{}, false},
		{":*", imap.NumRange{}, false},
		{"1:", imap.NumRange{}, false},
		{":1", imap.NumRange{}, false},
		{"0:0", imap.NumRange{}, false},
		{"0:*", imap.NumRange{}, false},
		{"0:1", imap.NumRange{}, false},
		{"1:0", imap.NumRange{}, false},
		{"1:2 ", imap.NumRange{}, false},
		{"1: 2", imap.NumRange{}, false},
		{"1:2:", imap.NumRange{}, false},
		{"1:2,", imap.NumRange{}, false},
		{"1:2:3", imap.NumRange{}, false},
		{"1:2,3", imap.NumRange{}, false},
		{"*:4294967296", imap.NumRange{}, false},
		{"0:4294967295", imap.NumRange{}, false},
		{"1:4294967296", imap.NumRange{}, false},
		{"4294967296:*", imap.NumRange{}, false},
		{"4294967295:0", imap.NumRange{}, false},
		{"4294967296:1", imap.NumRange{}, false},
		{"4294967295:4294967296", imap.NumRange{}, false},

		// Valid range
		{"*:*", imap.NumRange{0, 0}, true},
		{"1:*", imap.NumRange{1, 0}, true},
		{"*:1", imap.NumRange{1, 0}, true},
		{"2:2", imap.NumRange{2, 2}, true},
		{"2:42", imap.NumRange{2, 42}, true},
		{"42:2", imap.NumRange{2, 42}, true},
		{"*:4294967294", imap.NumRange{max - 1, 0}, true},
		{"*:4294967295", imap.NumRange{max, 0}, true},
		{"4294967294:*", imap.NumRange{max - 1, 0}, true},
		{"4294967295:*", imap.NumRange{max, 0}, true},
		{"1:4294967294", imap.NumRange{1, max - 1}, true},
		{"1:4294967295", imap.NumRange{1, max}, true},
		{"4294967295:1000", imap.NumRange{1000, max}, true},
		{"4294967294:4294967295", imap.NumRange{max - 1, max}, true},
		{"4294967295:4294967295", imap.NumRange{max, max}, true},
	}
	for _, test := range tests {
		out, err := parseNumRange(test.in)
		if !test.ok {
			if err == nil {
				t.Errorf("parseSeq(%q) expected error; got %q", test.in, out)
			}
		} else if err != nil {
			t.Errorf("parseSeq(%q) expected %q; got %v", test.in, test.out, err)
		} else if out != test.out {
			t.Errorf("parseSeq(%q) expected %q; got %q", test.in, test.out, out)
		}
	}
}

func TestNumRangeContainsLess(t *testing.T) {
	tests := []struct {
		s        string
		q        uint32
		contains bool
		less     bool
	}{
		{"2", 0, false, true},
		{"2", 1, false, false},
		{"2", 2, true, false},
		{"2", 3, false, true},
		{"2", max, false, true},

		{"*", 0, true, false},
		{"*", 1, false, false},
		{"*", 2, false, false},
		{"*", 3, false, false},
		{"*", max, false, false},

		{"2:3", 0, false, true},
		{"2:3", 1, false, false},
		{"2:3", 2, true, false},
		{"2:3", 3, true, false},
		{"2:3", 4, false, true},
		{"2:3", 5, false, true},

		{"2:4", 0, false, true},
		{"2:4", 1, false, false},
		{"2:4", 2, true, false},
		{"2:4", 3, true, false},
		{"2:4", 4, true, false},
		{"2:4", 5, false, true},

		{"4:4294967295", 0, false, true},
		{"4:4294967295", 1, false, false},
		{"4:4294967295", 2, false, false},
		{"4:4294967295", 3, false, false},
		{"4:4294967295", 4, true, false},
		{"4:4294967295", 5, true, false},
		{"4:4294967295", max, true, false},

		{"4:*", 0, true, false},
		{"4:*", 1, false, false},
		{"4:*", 2, false, false},
		{"4:*", 3, false, false},
		{"4:*", 4, true, false},
		{"4:*", 5, true, false},
		{"4:*", max, true, false},
	}
	for _, test := range tests {
		s, err := parseNumRange(test.s)
		if err != nil {
			t.Errorf("parseSeq(%q) unexpected error; %v", test.s, err)
			continue
		}
		if s.Contains(test.q) != test.contains {
			t.Errorf("%q.Contains(%d) expected %v", test.s, test.q, test.contains)
		}
		if s.Less(test.q) != test.less {
			t.Errorf("%q.Less(%d) expected %v", test.s, test.q, test.less)
		}
	}
}

func TestNumRangeMerge(T *testing.T) {
	tests := []struct {
		s, t, out string
	}{
		// Number with number
		{"1", "1", "1"},
		{"1", "2", "1:2"},
		{"1", "3", ""},
		{"1", "4294967295", ""},
		{"1", "*", ""},

		{"4", "1", ""},
		{"4", "2", ""},
		{"4", "3", "3:4"},
		{"4", "4", "4"},
		{"4", "5", "4:5"},
		{"4", "6", ""},

		{"4294967295", "4294967293", ""},
		{"4294967295", "4294967294", "4294967294:4294967295"},
		{"4294967295", "4294967295", "4294967295"},
		{"4294967295", "*", ""},

		{"*", "1", ""},
		{"*", "2", ""},
		{"*", "4294967294", ""},
		{"*", "4294967295", ""},
		{"*", "*", "*"},

		// NumRange with number
		{"1:3", "1", "1:3"},
		{"1:3", "2", "1:3"},
		{"1:3", "3", "1:3"},
		{"1:3", "4", "1:4"},
		{"1:3", "5", ""},
		{"1:3", "*", ""},

		{"3:4", "1", ""},
		{"3:4", "2", "2:4"},
		{"3:4", "3", "3:4"},
		{"3:4", "4", "3:4"},
		{"3:4", "5", "3:5"},
		{"3:4", "6", ""},
		{"3:4", "*", ""},

		{"2:3", "5", ""},
		{"2:4", "5", "2:5"},
		{"2:5", "5", "2:5"},
		{"2:6", "5", "2:6"},
		{"2:7", "5", "2:7"},
		{"2:*", "5", "2:*"},
		{"3:4", "5", "3:5"},
		{"3:5", "5", "3:5"},
		{"3:6", "5", "3:6"},
		{"3:7", "5", "3:7"},
		{"3:*", "5", "3:*"},
		{"4:5", "5", "4:5"},
		{"4:6", "5", "4:6"},
		{"4:7", "5", "4:7"},
		{"4:*", "5", "4:*"},
		{"5:6", "5", "5:6"},
		{"5:7", "5", "5:7"},
		{"5:*", "5", "5:*"},
		{"6:7", "5", "5:7"},
		{"6:*", "5", "5:*"},
		{"7:8", "5", ""},
		{"7:*", "5", ""},

		{"3:4294967294", "1", ""},
		{"3:4294967294", "2", "2:4294967294"},
		{"3:4294967294", "3", "3:4294967294"},
		{"3:4294967294", "4", "3:4294967294"},
		{"3:4294967294", "4294967293", "3:4294967294"},
		{"3:4294967294", "4294967294", "3:4294967294"},
		{"3:4294967294", "4294967295", "3:4294967295"},
		{"3:4294967294", "*", ""},

		{"3:4294967295", "1", ""},
		{"3:4294967295", "2", "2:4294967295"},
		{"3:4294967295", "3", "3:4294967295"},
		{"3:4294967295", "4", "3:4294967295"},
		{"3:4294967295", "4294967294", "3:4294967295"},
		{"3:4294967295", "4294967295", "3:4294967295"},
		{"3:4294967295", "*", ""},

		{"1:4294967295", "1", "1:4294967295"},
		{"1:4294967295", "4294967295", "1:4294967295"},
		{"1:4294967295", "*", ""},

		{"1:*", "1", "1:*"},
		{"1:*", "2", "1:*"},
		{"1:*", "4294967294", "1:*"},
		{"1:*", "4294967295", "1:*"},
		{"1:*", "*", "1:*"},

		// NumRange with range
		{"5:8", "1:2", ""},
		{"5:8", "1:3", ""},
		{"5:8", "1:4", "1:8"},
		{"5:8", "1:5", "1:8"},
		{"5:8", "1:6", "1:8"},
		{"5:8", "1:7", "1:8"},
		{"5:8", "1:8", "1:8"},
		{"5:8", "1:9", "1:9"},
		{"5:8", "1:10", "1:10"},
		{"5:8", "1:11", "1:11"},
		{"5:8", "1:*", "1:*"},

		{"5:8", "2:3", ""},
		{"5:8", "2:4", "2:8"},
		{"5:8", "2:5", "2:8"},
		{"5:8", "2:6", "2:8"},
		{"5:8", "2:7", "2:8"},
		{"5:8", "2:8", "2:8"},
		{"5:8", "2:9", "2:9"},
		{"5:8", "2:10", "2:10"},
		{"5:8", "2:11", "2:11"},
		{"5:8", "2:*", "2:*"},

		{"5:8", "3:4", "3:8"},
		{"5:8", "3:5", "3:8"},
		{"5:8", "3:6", "3:8"},
		{"5:8", "3:7", "3:8"},
		{"5:8", "3:8", "3:8"},
		{"5:8", "3:9", "3:9"},
		{"5:8", "3:10", "3:10"},
		{"5:8", "3:11", "3:11"},
		{"5:8", "3:*", "3:*"},

		{"5:8", "4:5", "4:8"},
		{"5:8", "4:6", "4:8"},
		{"5:8", "4:7", "4:8"},
		{"5:8", "4:8", "4:8"},
		{"5:8", "4:9", "4:9"},
		{"5:8", "4:10", "4:10"},
		{"5:8", "4:11", "4:11"},
		{"5:8", "4:*", "4:*"},

		{"5:8", "5:6", "5:8"},
		{"5:8", "5:7", "5:8"},
		{"5:8", "5:8", "5:8"},
		{"5:8", "5:9", "5:9"},
		{"5:8", "5:10", "5:10"},
		{"5:8", "5:11", "5:11"},
		{"5:8", "5:*", "5:*"},

		{"5:8", "6:7", "5:8"},
		{"5:8", "6:8", "5:8"},
		{"5:8", "6:9", "5:9"},
		{"5:8", "6:10", "5:10"},
		{"5:8", "6:11", "5:11"},
		{"5:8", "6:*", "5:*"},

		{"5:8", "7:8", "5:8"},
		{"5:8", "7:9", "5:9"},
		{"5:8", "7:10", "5:10"},
		{"5:8", "7:11", "5:11"},
		{"5:8", "7:*", "5:*"},

		{"5:8", "8:9", "5:9"},
		{"5:8", "8:10", "5:10"},
		{"5:8", "8:11", "5:11"},
		{"5:8", "8:*", "5:*"},

		{"5:8", "9:10", "5:10"},
		{"5:8", "9:11", "5:11"},
		{"5:8", "9:*", "5:*"},

		{"5:8", "10:11", ""},
		{"5:8", "10:*", ""},

		{"1:*", "1:*", "1:*"},
		{"1:*", "2:*", "1:*"},
		{"1:*", "1:4294967294", "1:*"},
		{"1:*", "1:4294967295", "1:*"},
		{"1:*", "2:4294967295", "1:*"},

		{"1:4294967295", "1:4294967294", "1:4294967295"},
		{"1:4294967295", "1:4294967295", "1:4294967295"},
		{"1:4294967295", "2:4294967295", "1:4294967295"},
		{"1:4294967295", "2:*", "1:*"},
	}
	for _, test := range tests {
		s, err := parseNumRange(test.s)
		if err != nil {
			T.Errorf("parseSeq(%q) unexpected error; %v", test.s, err)
			continue
		}
		t, err := parseNumRange(test.t)
		if err != nil {
			T.Errorf("parseSeq(%q) unexpected error; %v", test.t, err)
			continue
		}
		testOK := test.out != ""
		for i := 0; i < 2; i++ {
			if !testOK {
				test.out = test.s
			}
			out, ok := s.Merge(t)
			if out.String() != test.out || ok != testOK {
				T.Errorf("%q.Merge(%q) expected %q; got %q", test.s, test.t, test.out, out)
			}
			// Swap s & t, result should be identical
			test.s, test.t = test.t, test.s
			s, t = t, s
		}
	}
}

func checkNumSet(s imap.NumSet, t *testing.T) {
	n := len(s)
	for i, v := range s {
		if v.Start == 0 {
			if v.Stop != 0 {
				t.Errorf(`NumSet(%q) index %d: "*:n" range`, s, i)
			} else if i != n-1 {
				t.Errorf(`NumSet(%q) index %d: "*" not at the end`, s, i)
			}
			continue
		}
		if i > 0 && s[i-1].Stop >= v.Start-1 {
			t.Errorf(`NumSet(%q) index %d: overlap`, s, i)
		}
		if v.Stop < v.Start {
			if v.Stop != 0 {
				t.Errorf(`NumSet(%q) index %d: reversed range`, s, i)
			} else if i != n-1 {
				t.Errorf(`NumSet(%q) index %d: "n:*" not at the end`, s, i)
			}
		}
	}
}

func TestNumSetInfo(t *testing.T) {
	tests := []struct {
		s        string
		q        uint32
		contains bool
	}{
		{"", 0, false},
		{"", 1, false},
		{"", 2, false},
		{"", 3, false},
		{"", max, false},

		{"2", 0, false},
		{"2", 1, false},
		{"2", 2, true},
		{"2", 3, false},
		{"2", max, false},

		{"*", 0, false}, // Contains("*") is always false, use Dynamic() instead
		{"*", 1, false},
		{"*", 2, false},
		{"*", 3, false},
		{"*", max, false},

		{"1:*", 0, false},
		{"1:*", 1, true},
		{"1:*", max, true},

		{"2:4", 0, false},
		{"2:4", 1, false},
		{"2:4", 2, true},
		{"2:4", 3, true},
		{"2:4", 4, true},
		{"2:4", 5, false},
		{"2:4", max, false},

		{"2,4", 0, false},
		{"2,4", 1, false},
		{"2,4", 2, true},
		{"2,4", 3, false},
		{"2,4", 4, true},
		{"2,4", 5, false},
		{"2,4", max, false},

		{"2:4,6", 0, false},
		{"2:4,6", 1, false},
		{"2:4,6", 2, true},
		{"2:4,6", 3, true},
		{"2:4,6", 4, true},
		{"2:4,6", 5, false},
		{"2:4,6", 6, true},
		{"2:4,6", 7, false},

		{"2,4:6", 0, false},
		{"2,4:6", 1, false},
		{"2,4:6", 2, true},
		{"2,4:6", 3, false},
		{"2,4:6", 4, true},
		{"2,4:6", 5, true},
		{"2,4:6", 6, true},
		{"2,4:6", 7, false},

		{"2,4,6", 0, false},
		{"2,4,6", 1, false},
		{"2,4,6", 2, true},
		{"2,4,6", 3, false},
		{"2,4,6", 4, true},
		{"2,4,6", 5, false},
		{"2,4,6", 6, true},
		{"2,4,6", 7, false},

		{"1,3:5,7,9:*", 0, false},
		{"1,3:5,7,9:*", 1, true},
		{"1,3:5,7,9:*", 2, false},
		{"1,3:5,7,9:*", 3, true},
		{"1,3:5,7,9:*", 4, true},
		{"1,3:5,7,9:*", 5, true},
		{"1,3:5,7,9:*", 6, false},
		{"1,3:5,7,9:*", 7, true},
		{"1,3:5,7,9:*", 8, false},
		{"1,3:5,7,9:*", 9, true},
		{"1,3:5,7,9:*", 10, true},
		{"1,3:5,7,9:*", max, true},

		{"1,3:5,7,9,42", 0, false},
		{"1,3:5,7,9,42", 1, true},
		{"1,3:5,7,9,42", 2, false},
		{"1,3:5,7,9,42", 3, true},
		{"1,3:5,7,9,42", 4, true},
		{"1,3:5,7,9,42", 5, true},
		{"1,3:5,7,9,42", 6, false},
		{"1,3:5,7,9,42", 7, true},
		{"1,3:5,7,9,42", 8, false},
		{"1,3:5,7,9,42", 9, true},
		{"1,3:5,7,9,42", 10, false},
		{"1,3:5,7,9,42", 41, false},
		{"1,3:5,7,9,42", 42, true},
		{"1,3:5,7,9,42", 43, false},
		{"1,3:5,7,9,42", max, false},

		{"1,3:5,7,9,42,*", 0, false},
		{"1,3:5,7,9,42,*", 1, true},
		{"1,3:5,7,9,42,*", 2, false},
		{"1,3:5,7,9,42,*", 3, true},
		{"1,3:5,7,9,42,*", 4, true},
		{"1,3:5,7,9,42,*", 5, true},
		{"1,3:5,7,9,42,*", 6, false},
		{"1,3:5,7,9,42,*", 7, true},
		{"1,3:5,7,9,42,*", 8, false},
		{"1,3:5,7,9,42,*", 9, true},
		{"1,3:5,7,9,42,*", 10, false},
		{"1,3:5,7,9,42,*", 41, false},
		{"1,3:5,7,9,42,*", 42, true},
		{"1,3:5,7,9,42,*", 43, false},
		{"1,3:5,7,9,42,*", max, false},

		{"1,3:5,7,9,42,60:70,100:*", 0, false},
		{"1,3:5,7,9,42,60:70,100:*", 1, true},
		{"1,3:5,7,9,42,60:70,100:*", 2, false},
		{"1,3:5,7,9,42,60:70,100:*", 3, true},
		{"1,3:5,7,9,42,60:70,100:*", 4, true},
		{"1,3:5,7,9,42,60:70,100:*", 5, true},
		{"1,3:5,7,9,42,60:70,100:*", 6, false},
		{"1,3:5,7,9,42,60:70,100:*", 7, true},
		{"1,3:5,7,9,42,60:70,100:*", 8, false},
		{"1,3:5,7,9,42,60:70,100:*", 9, true},
		{"1,3:5,7,9,42,60:70,100:*", 10, false},
		{"1,3:5,7,9,42,60:70,100:*", 41, false},
		{"1,3:5,7,9,42,60:70,100:*", 42, true},
		{"1,3:5,7,9,42,60:70,100:*", 43, false},
		{"1,3:5,7,9,42,60:70,100:*", 59, false},
		{"1,3:5,7,9,42,60:70,100:*", 60, true},
		{"1,3:5,7,9,42,60:70,100:*", 65, true},
		{"1,3:5,7,9,42,60:70,100:*", 70, true},
		{"1,3:5,7,9,42,60:70,100:*", 71, false},
		{"1,3:5,7,9,42,60:70,100:*", 99, false},
		{"1,3:5,7,9,42,60:70,100:*", 100, true},
		{"1,3:5,7,9,42,60:70,100:*", 1000, true},
		{"1,3:5,7,9,42,60:70,100:*", max, true},
	}
	for _, test := range tests {
		s, _ := ParseNumSet(test.s)
		checkNumSet(s, t)
		if s.Contains(test.q) != test.contains {
			t.Errorf("%q.Contains(%v) expected %v", test.s, test.q, test.contains)
		}
		if str := s.String(); str != test.s {
			t.Errorf("%q.String() expected %q; got %q", test.s, test.s, str)
		}
		testEmpty := len(test.s) == 0
		if (len(s) == 0) != testEmpty {
			t.Errorf("%q.Empty() expected %v", test.s, testEmpty)
		}
		testDynamic := !testEmpty && test.s[len(test.s)-1] == '*'
		if s.Dynamic() != testDynamic {
			t.Errorf("%q.Dynamic() expected %v", test.s, testDynamic)
		}
	}
}

func TestParseNumSet(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"1,1", "1"},
		{"1,2", "1:2"},
		{"1,3", "1,3"},
		{"1,*", "1,*"},

		{"1,1,1", "1"},
		{"1,1,2", "1:2"},
		{"1,1:2", "1:2"},
		{"1,1,3", "1,3"},
		{"1,1:3", "1:3"},
		{"1,2,2", "1:2"},
		{"1,2,3", "1:3"},
		{"1,2:3", "1:3"},
		{"1,2,4", "1:2,4"},
		{"1,3,3", "1,3"},
		{"1,3,4", "1,3:4"},
		{"1,3:4", "1,3:4"},
		{"1,3,5", "1,3,5"},
		{"1,3:5", "1,3:5"},
		{"1:3,5", "1:3,5"},
		{"1:5,3", "1:5"},

		{"1,2,3,4", "1:4"},
		{"1,2,4,5", "1:2,4:5"},
		{"1,2,4:5", "1:2,4:5"},
		{"1:2,4:5", "1:2,4:5"},

		{"1,2,3,4,5", "1:5"},
		{"1,2:3,4:5", "1:5"},

		{"1,2,4,5,7,9", "1:2,4:5,7,9"},
		{"1,2,4,5,7:9", "1:2,4:5,7:9"},
		{"1:2,4:5,7:9", "1:2,4:5,7:9"},
		{"1,2,4,5,7,8,9", "1:2,4:5,7:9"},
		{"1:2,4:5,7,8,9", "1:2,4:5,7:9"},

		{"3,5:10,15:20", "3,5:10,15:20"},
		{"4,5:10,15:20", "4:10,15:20"},
		{"5,5:10,15:20", "5:10,15:20"},
		{"7,5:10,15:20", "5:10,15:20"},
		{"10,5:10,15:20", "5:10,15:20"},
		{"11,5:10,15:20", "5:11,15:20"},
		{"12,5:10,15:20", "5:10,12,15:20"},
		{"14,5:10,15:20", "5:10,14:20"},
		{"17,5:10,15:20", "5:10,15:20"},
		{"21,5:10,15:20", "5:10,15:21"},
		{"22,5:10,15:20", "5:10,15:20,22"},
		{"*,5:10,15:20", "5:10,15:20,*"},

		{"1:3,5:10,15:20", "1:3,5:10,15:20"},
		{"1:4,5:10,15:20", "1:10,15:20"},
		{"1:8,5:10,15:20", "1:10,15:20"},
		{"1:13,5:10,15:20", "1:13,15:20"},
		{"1:14,5:10,15:20", "1:20"},
		{"7:17,5:10,15:20", "5:20"},
		{"11:14,5:10,15:20", "5:20"},
		{"12,13,5:10,15:20", "5:10,12:13,15:20"},
		{"12:13,5:10,15:20", "5:10,12:13,15:20"},
		{"12:14,5:10,15:20", "5:10,12:20"},
		{"11:13,5:10,15:20", "5:13,15:20"},
		{"11,12,13,14,5:10,15:20", "5:20"},

		{"1:*,5:10,15:20", "1:*"},
		{"4:*,5:10,15:20", "4:*"},
		{"6:*,5:10,15:20", "5:*"},
		{"12:*,5:10,15:20", "5:10,12:*"},
		{"19:*,5:10,15:20", "5:10,15:*"},

		{"5:8,6,7:10,15,16,17,18:20,19,21:*", "5:10,15:*"},

		{"4:13,1,5,10,15,20", "1,4:13,15,20"},
		{"4:14,1,5,10,15,20", "1,4:15,20"},
		{"4:15,1,5,10,15,20", "1,4:15,20"},
		{"4:16,1,5,10,15,20", "1,4:16,20"},
		{"4:17,1,5,10,15,20", "1,4:17,20"},
		{"4:18,1,5,10,15,20", "1,4:18,20"},
		{"4:19,1,5,10,15,20", "1,4:20"},
		{"4:20,1,5,10,15,20", "1,4:20"},
		{"4:21,1,5,10,15,20", "1,4:21"},
		{"4:*,1,5,10,15,20", "1,4:*"},

		{"1,3,5,7,9,11,13,15,17,19", "1,3,5,7,9,11,13,15,17,19"},
		{"1,3,5,7,9,11:13,15,17,19", "1,3,5,7,9,11:13,15,17,19"},
		{"1,3,5,7,9:11,13:15,17,19", "1,3,5,7,9:11,13:15,17,19"},
		{"1,3,5,7:9,11:13,15:17,19", "1,3,5,7:9,11:13,15:17,19"},
		{"1,3,5,7,9,11,13,15,17,19,*", "1,3,5,7,9,11,13,15,17,19,*"},
		{"1,3,5,7,9,11,13,15,17,19:*", "1,3,5,7,9,11,13,15,17,19:*"},
		{"1:20,3,5,7,9,11,13,15,17,19,*", "1:20,*"},
		{"1:20,3,5,7,9,11,13,15,17,19:*", "1:*"},

		{"4294967295,*", "4294967295,*"},
		{"1,4294967295,*", "1,4294967295,*"},
		{"1:4294967295,*", "1:4294967295,*"},
		{"1,4294967295:*", "1,4294967295:*"},
		{"1:*,4294967295", "1:*"},
		{"1:*,4294967295:*", "1:*"},
		{"1:4294967295,4294967295:*", "1:*"},
	}
	prng := rand.New(rand.NewSource(19860201))
	done := make(map[string]bool)
	permute := func(in string) string {
		v := strings.Split(in, ",")
		r := make([]string, len(v))

		// Try to find a permutation that hasn't been checked already
		for i := 0; i < 50; i++ {
			for i, j := range prng.Perm(len(v)) {
				r[i] = v[j]
			}
			if s := strings.Join(r, ","); !done[s] {
				done[s] = true
				return s
			}
		}
		return ""
	}
	for _, test := range tests {
		for i := 0; i < 100 && test.in != ""; i++ {
			s, err := ParseNumSet(test.in)
			if err != nil {
				t.Errorf("Add(%q) unexpected error; %v", test.in, err)
				i = 100
			}
			checkNumSet(s, t)
			if out := s.String(); out != test.out {
				t.Errorf("%q.String() expected %q; got %q", test.in, test.out, out)
				i = 100
			}
			test.in = permute(test.in)
		}
	}
}

func TestNumSetAddNumRangeSet(t *testing.T) {
	type num []uint32
	tests := []struct {
		num num
		rng imap.NumRange
		set string
		out string
	}{
		{num{5}, imap.NumRange{1, 3}, "1:2,5,7:13,15,17:*", "1:3,5,7:13,15,17:*"},
		{num{5}, imap.NumRange{3, 1}, "2:3,7:13,15,17:*", "1:3,5,7:13,15,17:*"},

		{num{15}, imap.NumRange{17, 0}, "1:3,5,7:13", "1:3,5,7:13,15,17:*"},
		{num{15}, imap.NumRange{0, 17}, "1:3,5,7:13", "1:3,5,7:13,15,17:*"},

		{num{1, 3, 5, 7, 9, 11, 0}, imap.NumRange{8, 13}, "2,15,17:*", "1:3,5,7:13,15,17:*"},
		{num{5, 1, 7, 3, 9, 0, 11}, imap.NumRange{8, 13}, "2,15,17:*", "1:3,5,7:13,15,17:*"},
		{num{5, 1, 7, 3, 9, 0, 11}, imap.NumRange{13, 8}, "2,15,17:*", "1:3,5,7:13,15,17:*"},
	}
	for _, test := range tests {
		other, _ := ParseNumSet(test.set)

		var s imap.NumSet
		s.AddNum(test.num...)
		checkNumSet(s, t)
		s.AddRange(test.rng.Start, test.rng.Stop)
		checkNumSet(s, t)
		s.AddSet(other)
		checkNumSet(s, t)

		if out := s.String(); out != test.out {
			t.Errorf("(%v + %v + %q).String() expected %q; got %q", test.num, test.rng, test.set, test.out, out)
		}
	}
}
