package anscombe

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

// TestWriteSVG checks that the plot is well-formed XML with one point per
// observation, one fitted line and one title per data set.
func TestWriteSVG(t *testing.T) {
	sets := Quartet()
	var buf bytes.Buffer
	if err := WriteSVG(&buf, sets); err != nil {
		t.Fatal(err)
	}

	decoder := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
	counts := map[string]int{}
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("SVG is not valid XML: %v", err)
		}
		if start, ok := tok.(xml.StartElement); ok {
			counts[start.Name.Local]++
		}
	}

	if counts["circle"] != 44 {
		t.Errorf("found %d points, want 44 (4 sets x 11)", counts["circle"])
	}
	if counts["line"] != 4 {
		t.Errorf("found %d fitted lines, want 4", counts["line"])
	}
	for _, ds := range sets {
		if !strings.Contains(buf.String(), ">"+ds.Name+"<") {
			t.Errorf("missing title %q", ds.Name)
		}
	}
}

func TestWriteSVGRejectsBadInput(t *testing.T) {
	if err := WriteSVG(&bytes.Buffer{}, nil); err == nil {
		t.Error("expected an error for no data sets")
	}
	flat := []DataSet{{Name: "flat", X: []float64{1, 1, 1}, Y: []float64{1, 2, 3}}}
	if err := WriteSVG(&bytes.Buffer{}, flat); err == nil {
		t.Error("expected an error for constant x")
	}
}
