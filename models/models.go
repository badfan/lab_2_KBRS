package models

type LogInInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	N        string `json:"n"`
	E        string `json:"e"`
}

type CreateUpdateFileInput struct {
	FileName string `json:"file_name"`
	Text     string `json:"text"`
}

type GetDeleteFileInput struct {
	FileName string `json:"file_name"`
}
