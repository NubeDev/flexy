package code

const (
	SUCCESS        = 200
	ERROR          = 500
	InvalidParams  = 400
	TokenInvalid   = 401
	UnknownError   = 900
	UnknownCommand = 902

	ErrorAuthCheckTokenFail     = 20001
	ErrorAuthCheckTokenTimeout  = 20002
	ErrorAuthToken              = 20003
	ErrorAuth                   = 20004
	ErrorUserPasswordInvalid    = 20005
	AuthTokenInBlockList        = 20006
	ErrorUserOldPasswordInvalid = 20007
	AccessTokenFailure          = 20008
	RefreshAccessTokenFailure   = 20009
)

var MsgFlags = map[int]string{
	SUCCESS:                     "ok",
	ERROR:                       "fail",
	UnknownError:                "Unknown error",
	UnknownCommand:              "Unknown command",
	InvalidParams:               "Request parameter error",
	TokenInvalid:                "Token parameter is invalid or does not exist",
	ErrorAuthCheckTokenFail:     "Token authorization failed",
	ErrorAuthCheckTokenTimeout:  "Token has expired",
	ErrorAuthToken:              "Token generation failed",
	ErrorAuth:                   "Token error",
	ErrorUserPasswordInvalid:    "Incorrect username or password",
	AuthTokenInBlockList:        "The token is already in the block list",
	ErrorUserOldPasswordInvalid: "Incorrect original password for the user",
	AccessTokenFailure:          "Access token generation failed",
	RefreshAccessTokenFailure:   "Refresh token generation failed",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
