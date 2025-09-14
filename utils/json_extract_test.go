package utils

import (
    "encoding/json"
    "testing"
)

func TestExtractJSONFromText_ObjectOnly(t *testing.T) {
    in := `{"a":1,"b":[2,3],"c":{"d":"e"}}`
    out, err := ExtractJSONFromText(in)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if out != in {
        t.Fatalf("expected identical output, got: %s", out)
    }
}

func TestExtractJSONFromText_WithPreambleAndSuffix(t *testing.T) {
    in := "Here is the result:\n```json\n{\n  \"x\": [1, 2, 3],\n  \"y\": \"ok\"\n}\n```\nThanks!"
    out, err := ExtractJSONFromText(in)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal([]byte(out), &m); err != nil {
        t.Fatalf("unmarshal failed: %v\njson: %s", err, out)
    }
    if m["y"].(string) != "ok" {
        t.Fatalf("unexpected content: %#v", m)
    }
}

func TestExtractJSONFromText_Array(t *testing.T) {
    in := "prefix [ {\n\"a\":1}, {\"b\":2} ] suffix"
    out, err := ExtractJSONFromText(in)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    var v []map[string]int
    if err := json.Unmarshal([]byte(out), &v); err != nil {
        t.Fatalf("unmarshal failed: %v\njson: %s", err, out)
    }
    if len(v) != 2 || v[1]["b"] != 2 {
        t.Fatalf("unexpected content: %#v", v)
    }
}

func TestExtractJSONFromText_BracesInsideString(t *testing.T) {
    in := "noise {\n  \"text\": \"closing } inside\",\n  \"ok\":true\n} tail"
    out, err := ExtractJSONFromText(in)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    var m map[string]any
    if err := json.Unmarshal([]byte(out), &m); err != nil {
        t.Fatalf("unmarshal failed: %v\njson: %s", err, out)
    }
    if m["ok"].(bool) != true {
        t.Fatalf("unexpected content: %#v", m)
    }
}

func TestExtractJSONFromText_NoJSON(t *testing.T) {
    in := "No JSON here, sorry."
    if _, err := ExtractJSONFromText(in); err == nil {
        t.Fatalf("expected error, got none")
    }
}

