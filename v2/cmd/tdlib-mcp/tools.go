package main

// toolSpec is one entry of the tools/list result.
type toolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// obj is a shorthand for building JSON Schema fragments.
func obj(props map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func numProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}

var tools = []toolSpec{
	{
		Name: "tdlib_search_methods",
		Description: "Search the TDLib method registry by name or description. " +
			"Returns matching methods with trimmed descriptions, parameter names/Go types and return types.",
		InputSchema: obj(map[string]any{
			"query": strProp("Case-insensitive substring matched against method names and descriptions."),
			"limit": numProp("Max results (default 10, max 50)."),
		}, "query"),
	},
	{
		Name: "tdlib_describe_method",
		Description: "Return the full schema of one TDLib method: every parameter with its description, " +
			"Go type and, for union parameters, the allowed \"@type\" variant constructors.",
		InputSchema: obj(map[string]any{
			"method": strProp(`TDLib method name, e.g. "sendMessage".`),
		}, "method"),
	},
	{
		Name: "tdlib_call",
		Description: "Call a TDLib method on the live session and return the raw JSON response. " +
			"Params are validated against the method schema first; nested objects keep their \"@type\" fields. " +
			"TDLib errors are returned as {code, message} with isError=true.",
		InputSchema: obj(map[string]any{
			"method": strProp(`TDLib method name, e.g. "getChats".`),
			"params": map[string]any{
				"type":                 "object",
				"description":          `Method parameters keyed by their snake_case JSON names. Do not include a top-level "@type"; it is added automatically.`,
				"additionalProperties": true,
			},
		}, "method"),
	},
	{
		Name: "tdlib_get_updates",
		Description: "Return recent TDLib updates from the in-memory ring buffer (newest last), " +
			"optionally filtered by @type.",
		InputSchema: obj(map[string]any{
			"type":  strProp(`Optional exact @type filter, e.g. "updateNewMessage".`),
			"limit": numProp("Max updates to return (default 50)."),
		}),
	},
	{
		Name: "tdlib_wait_update",
		Description: "Block until a TDLib update matching the filters arrives, or the timeout elapses. " +
			"Useful after tdlib_call to await the resulting update.",
		InputSchema: obj(map[string]any{
			"type":            strProp(`Optional exact @type filter, e.g. "updateNewMessage".`),
			"timeout_seconds": numProp("Seconds to wait (default 30, max 120)."),
			"contains":        strProp("Optional substring that the raw update JSON must contain."),
		}),
	},
	{
		Name:        "tdlib_auth_status",
		Description: "Return the current TDLib authorization state kind and whether the session is fully authorized.",
		InputSchema: obj(map[string]any{}),
	},
	{
		Name: "tdlib_submit_code",
		Description: "Submit an authentication code while the login flow is waiting for one " +
			"(authorizationStateWaitCode). Check tdlib_auth_status first.",
		InputSchema: obj(map[string]any{
			"code": strProp("The login code received via Telegram."),
		}, "code"),
	},
}

func toolSpecs() []toolSpec { return tools }

func knownTool(name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}
