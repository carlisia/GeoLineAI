package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"geolineai/internal/core/ids"
	"geolineai/internal/core/review/verdict"
)

// maxBody is the largest request body accepted.
const maxBody = 4 << 20

// readObject reads a body that must be a JSON object, answering a body over the
// limit and one of another shape; every object body shares those two answers. It
// returns the object's members and the names given more than once.
func readObject(c *gin.Context) (map[string]json.RawMessage, []string, bool) {
	raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, maxBody))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(c, http.StatusRequestEntityTooLarge, ErrorBody{Error: CodeBodyTooLarge, Message: fmt.Sprintf("The request body exceeds %d MiB.", maxBody>>20)})
		return nil, nil, false
	}
	if respondIfFailed(c, err) {
		return nil, nil, false
	}
	body, repeated, ok := object(raw)
	if !ok {
		validationError(c, "The body must be a JSON object.")
		return nil, nil, false
	}
	return body, repeated, true
}

// decodeSubmission decodes a save body's shape: a JSON object with only
// targets and verdict, each target an object with only a string reading_id.
// It reports every shape error at once; the rules about the targets' count,
// their IDs' form and the verdict's value are verdict's, checked afterwards.
func decodeSubmission(c *gin.Context) (verdict.Submission, bool) {
	var problems []string
	var sub verdict.Submission
	body, repeated, ok := readObject(c)
	if !ok {
		return sub, false
	}
	problems = append(problems, unknown(body, "", "targets", "verdict")...)
	problems = append(problems, givenOnce("", repeated)...)

	if t, ok := body["targets"]; !ok {
		problems = append(problems, "targets is required.")
	} else {
		var targets []json.RawMessage
		if json.Unmarshal(t, &targets) != nil || targets == nil {
			problems = append(problems, "targets must be an array.")
		}
		for i, rawTarget := range targets {
			path := fmt.Sprintf("targets[%d]", i)
			target, repeated, ok := object(rawTarget)
			if !ok {
				problems = append(problems, path+" must be an object.")
				continue
			}
			problems = append(problems, unknown(target, path+".", "reading_id")...)
			problems = append(problems, givenOnce(path+".", repeated)...)
			rid, ok := target["reading_id"]
			if !ok {
				problems = append(problems, path+".reading_id is required.")
				continue
			}
			var id string
			if isNull(rid) || json.Unmarshal(rid, &id) != nil {
				problems = append(problems, path+".reading_id "+ids.ReadingIDRule)
				continue
			}
			sub.Targets = append(sub.Targets, verdict.Target{Index: i, ReadingID: id})
		}
	}

	var problem string
	if sub.Verdict, problem = stringMember(body, "verdict"); problem != "" {
		problems = append(problems, problem)
	}

	if len(problems) > 0 {
		validationError(c, problems...)
		return sub, false
	}
	return sub, true
}

// decodeStrings decodes a body that must be a JSON object holding exactly the named
// members, each a string, reporting every shape error at once; the rules about the
// strings' contents are the caller's, checked afterwards.
func decodeStrings(c *gin.Context, names ...string) (map[string]string, bool) {
	body, repeated, ok := readObject(c)
	if !ok {
		return nil, false
	}
	problems := append(unknown(body, "", names...), givenOnce("", repeated)...)
	out := make(map[string]string, len(names))
	for _, name := range names {
		value, problem := stringMember(body, name)
		if problem != "" {
			problems = append(problems, problem)
			continue
		}
		out[name] = value
	}
	if len(problems) > 0 {
		validationError(c, problems...)
		return nil, false
	}
	return out, true
}

// stringMember reads a required string member of a body, or the sentence naming the
// rule it breaks.
func stringMember(body map[string]json.RawMessage, name string) (string, string) {
	raw, ok := body[name]
	if !ok {
		return "", name + " is required."
	}
	var value string
	if isNull(raw) || json.Unmarshal(raw, &value) != nil {
		return "", name + " must be a string."
	}
	return value, ""
}

// isNull reports a JSON null, which encoding/json decodes into a string as "" with
// no error, so the shape check has to look for it.
func isNull(raw json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }

// object decodes a JSON object's raw members, naming any repeated.
func object(raw []byte) (map[string]json.RawMessage, []string, bool) {
	// encoding/json keeps only the last of a repeated member, so walk the tokens.
	dec := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, nil, false
	}
	m := map[string]json.RawMessage{}
	var repeated []string
	for dec.More() {
		tok, err := dec.Token()
		key, isKey := tok.(string)
		var value json.RawMessage
		if err != nil || !isKey || dec.Decode(&value) != nil {
			return nil, nil, false
		}
		if _, seen := m[key]; seen && !slices.Contains(repeated, key) {
			repeated = append(repeated, key)
		}
		m[key] = value
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return nil, nil, false
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, nil, false
	}
	slices.Sort(repeated)
	return m, repeated, true
}

// givenOnce names every member given more than once.
func givenOnce(prefix string, names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = prefix + n + " must be given once."
	}
	return out
}

// unknown names every member of m that is not allowed.
func unknown(m map[string]json.RawMessage, prefix string, allowed ...string) []string {
	var out []string
	for k := range m {
		if !slices.Contains(allowed, k) {
			out = append(out, fmt.Sprintf("%s%s is not a known field.", prefix, k))
		}
	}
	slices.Sort(out)
	return out
}
