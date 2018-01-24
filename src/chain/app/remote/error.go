package remote

import "fmt"

const (
	REMOTE_APP_LOG = "[remote_app]"
)

var APP_ERR = fmt.Errorf("app error")

type RemoteAppError struct {
	CONN_ALREADY_EXIST string
	CONN_TLS_FAILED    string
	CONN_DAIL_FAILED   string
	PB_MARSHAL_ERR     string
	PB_UNMARSHAL_ERR   string
	SEND_TOREMOTE_ERR  string
	EXECUTE_REMOTE_ERR string
}

func (e *RemoteAppError) init() {
	e.CONN_ALREADY_EXIST = fmt.Sprintf("%v,connection already exist", REMOTE_APP_LOG)
	e.CONN_TLS_FAILED = fmt.Sprintf("%v,Failed to create TLS credentials", REMOTE_APP_LOG)
	e.CONN_DAIL_FAILED = fmt.Sprintf("%v,fail to dial", REMOTE_APP_LOG)
	e.PB_MARSHAL_ERR = fmt.Sprintf("%v,proto marshal err", REMOTE_APP_LOG)
	e.PB_UNMARSHAL_ERR = fmt.Sprintf("%v,proto unmarshal err", REMOTE_APP_LOG)
	e.SEND_TOREMOTE_ERR = fmt.Sprintf("%v,send to remote err", REMOTE_APP_LOG)
	e.EXECUTE_REMOTE_ERR = fmt.Sprintf("%v,execute remote ret err", REMOTE_APP_LOG)
}

var rError RemoteAppError

func init() {
	rError.init()
}
