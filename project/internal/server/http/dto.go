package http

// представляет запрос на проверку попытки авторизации.
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

// представляет ответ на проверку авторизации.
type AuthResponse struct {
	OK bool `json:"ok"`
}

// представляет запрос на сброс бакета.
type ResetRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

// представляет запрос на добавление/удаление подсети в black/white-списки.
type CidrRequest struct {
	Subnet string `json:"subnet"`
}

// возвращает стандартную ошибку в формате JSON.
type ErrorResponse struct {
	Error string `json:"error"`
}
