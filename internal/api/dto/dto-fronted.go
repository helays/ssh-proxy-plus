package dto

type FrontedResp struct {
	Path      string         `json:"path"`
	Name      string         `json:"name"`
	Mod       string         `json:"mod,omitempty"`
	Component string         `json:"component"`
	Meta      map[string]any `json:"meta"`
}

type WsResp struct {
	Action string `json:"action"`
	Data   any    `json:"data"`
}
