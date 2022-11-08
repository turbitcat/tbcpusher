package api

type ApiEntry struct {
	Api          string   `json:"api:omitempty"`
	StringParams []string `json:"string_params:omitempty"`
	OtherParams  []string `json:"other_params:omitempty"`
	ReturnValue  any      `json:"return_value:omitempty"`
}

var Docs []ApiEntry = []ApiEntry{
	{Api: "/group/create", OtherParams: []string{"data"}, ReturnValue: "group id"},
	{Api: "/group/push", StringParams: []string{"group", "author", "title", "content", "when"}, OtherParams: []string{"ids of pushed sessions"}},
	{Api: "/group/setdata", StringParams: []string{"group"}, OtherParams: []string{"data"}, ReturnValue: "empty"},
	{Api: "/session/create", StringParams: []string{"group", "hook", "data"}, ReturnValue: "session id"},
	{Api: "/session/push", StringParams: []string{"session", "author", "title", "content", "when"}, ReturnValue: "empty"},
	{Api: "/session/check", StringParams: []string{"session"}, ReturnValue: "session info"},
	{Api: "/session/setdata", StringParams: []string{"session"}, OtherParams: []string{"data"}, ReturnValue: "empty"},
	{Api: "/session/hide", StringParams: []string{"session"}, ReturnValue: "empty"},
}
