package dto

type LoginReq struct {
	Pass    string `json:"pass"`
	Captcha string `json:"captcha"`
}
