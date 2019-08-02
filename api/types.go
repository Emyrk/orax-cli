package api

type RegisterUserResult struct {
	ID  string `json:"id"`
	JWT string `json:"jwt"`
}

type AuthenticateResult struct {
	ID  string `json:"id"`
	JWT string `json:"jwt"`
}

type RegisterMinerResult struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type Error struct {
	Message string `json:"error"`
	Code    int    `json:"code"`
}
