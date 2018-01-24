package remote

import (
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	ethcmn "github.com/Baptist-Publication/chorus-module/lib/eth/common"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	INVOKE_TIME_UP = time.Minute * 1
)

var (
	hashnull ethcmn.Hash
	accnull  ethcmn.Address
)

type RemoteAppConfig struct {
	tls_on               bool
	ca_file              string
	server_host_override string
	rpcapp_addr          string
	server_addr          string
}

func (c *RemoteAppConfig) Init(config *viper.Viper) {
	c.tls_on = config.GetBool("tls_on")
	c.ca_file = config.GetString("ca_file")
	c.server_host_override = config.GetString("server_host_override")
	c.server_addr = config.GetString("rpcapp_addr")
}

func (c *RemoteAppConfig) TLSOn() bool {
	return c.tls_on
}

func (c *RemoteAppConfig) CAFilePath() string {
	return c.ca_file
}

func (c *RemoteAppConfig) ServerHostOverride() string {
	return c.server_host_override
}

func (c *RemoteAppConfig) ServerAddr() string {
	return c.server_addr
}

type RemoteAppClient struct {
	conf RemoteAppConfig
	conn *grpc.ClientConn
	app  RemoteAppServiceClient
	log  *zap.Logger
}

func (client *RemoteAppClient) Init(log *zap.Logger, config *viper.Viper) bool {
	client.log = log
	client.conf.Init(config)
	if !client.makeConn() {
		return false
	}
	client.app = NewRemoteAppServiceClient(client.conn)
	return true
}

// TODO
func (client *RemoteAppClient) OnStart() {
}

func (client *RemoteAppClient) OnStop() {
	client.conn.Close()
}

func (client *RemoteAppClient) makeConn() bool {
	if client.conn != nil {
		client.log.Warn(rError.CONN_ALREADY_EXIST)
		return false
	}
	var opts []grpc.DialOption
	if client.conf.TLSOn() {
		creds, err := credentials.NewClientTLSFromFile(client.conf.CAFilePath(), client.conf.ServerHostOverride())
		if err != nil {
			client.log.Warn(rError.CONN_TLS_FAILED, zap.Error(err))
			return false
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(client.conf.ServerAddr(), opts...)
	if err != nil {
		client.log.Warn(rError.CONN_DAIL_FAILED, zap.Error(err))
		return false
	}
	client.conn = conn
	return true
}

func (client *RemoteAppClient) CheckTx(account ethcmn.Address, txdata []byte) error {
	cmd := ClientCheckCmd{
		AccountID: account.Hex(),
		Tx:        txdata,
	}
	res, err := client.app.CheckTx(context.Background(), &cmd)
	if err != nil {
		client.log.Warn(rError.EXECUTE_REMOTE_ERR, zap.Error(err))
		return err
	}
	if res.GetResult().GetCode() != CodeType_OK {
		return fmt.Errorf(res.GetResult().GetLog())
	}
	return nil
}

func (client *RemoteAppClient) ApplyTransaction(state *stateDup, account ethcmn.Address, txdata []byte) ([]byte, error) {
	do, err := client.app.Execute(context.Background())
	if err != nil {
		client.log.Warn(rError.EXECUTE_REMOTE_ERR, zap.Error(err))
		return nil, err
	}
	var revErr, sendErr error
	var rev *ServerCmd
	var timeUp bool
	done := make(chan bool, 0)
	go func() {
		timer := time.NewTimer(INVOKE_TIME_UP)
		select {
		case <-timer.C:
			timeUp = true
		case <-done:
		}
	}()
	accid := account.Hex()
	txCmd := client.GenTxCmd(accid, txdata, int64(state.height))
	if sendErr = do.Send(&txCmd); err != nil {
		client.log.Error(rError.SEND_TOREMOTE_ERR, zap.Error(sendErr))
		return nil, err
	}
	var receipt []byte
	for {
		// time may related to fee
		rev, revErr = do.Recv()
		if revErr == io.EOF || timeUp {
			close(done)
			break
		}
		send, brk := client.ServeCmd(state, accid, rev)
		if brk {
			receipt = rev.GetData()
			break
		}
		if sendErr = do.Send(&send); err != nil {
			client.log.Error(rError.SEND_TOREMOTE_ERR, zap.Error(sendErr))
			return nil, err
		}
	}
	do.CloseSend()
	if timeUp {
		return nil, fmt.Errorf("invoke time out")
	}
	return receipt, nil
}

func (client *RemoteAppClient) GenCmd(tp C_CmdType, accid string, data []byte) (cmd ClientCmd) {
	cmd.Type = tp
	cmd.AccountID = accid
	cmd.Data = data
	return
}

func (client *RemoteAppClient) GenTxCmd(accid string, txData []byte, blockHeight int64) (cmd ClientCmd) {
	cmd.Type = C_CmdType_Tx
	cmd.AccountID = accid
	txcmd := C_TxCmd{}
	txcmd.Tx = txData
	txcmd.BlockHeight = blockHeight
	cmd.Data = client.ToBytes(&txcmd)
	return
}

func (client *RemoteAppClient) ToBytes(msg proto.Message) []byte {
	ret, err := proto.Marshal(msg)
	if err != nil {
		client.log.Warn(rError.PB_MARSHAL_ERR, zap.Error(err))
		return nil
	}
	return ret
}

func (client *RemoteAppClient) ToMsg(data []byte, msg proto.Message) bool {
	err := proto.Unmarshal(data, msg)
	if err != nil {
		client.log.Warn(rError.PB_UNMARSHAL_ERR, zap.Error(err))
		return false
	}
	return true
}

func (client *RemoteAppClient) ServeCmd(state *stateDup, accID string, scmd *ServerCmd) (ret ClientCmd, brk bool) {
	if scmd == nil || len(scmd.Data) == 0 {
		ret = client.CmdRetResult(CodeType_MsgMissParam, nil, "message data is nil")
		return
	}
	defer func() {
		ret.Hash = scmd.Hash
	}()
	switch scmd.Type {
	case S_CmdType_AddData:
		var cmd S_AddDataCmd
		client.ToMsg(scmd.Data, &cmd)
		if len(cmd.Key) == 0 || len(cmd.Value) == 0 {
			ret = client.CmdRetResult(CodeType_MsgMissParam, nil, "key or value is nil")
			return
		}
		return client.doAddData(state, accID, cmd.Key, cmd.Value)
	case S_CmdType_DelData:
		var cmd S_DelDataCmd
		client.ToMsg(scmd.Data, &cmd)
		if len(cmd.Key) == 0 {
			ret = client.CmdRetResult(CodeType_MsgMissParam, nil, "key is nil")
			return
		}
		return client.doDelData(state, accID, cmd.Key)
	case S_CmdType_ModifyData:
		var cmd S_ModifyDataCmd
		client.ToMsg(scmd.Data, &cmd)
		if len(cmd.Key) == 0 || len(cmd.Value) == 0 {
			ret = client.CmdRetResult(CodeType_MsgMissParam, nil, "key or value is nil")
			return
		}
		return client.doDelData(state, accID, cmd.Key)
	case S_CmdType_QueryData:
		var cmd S_ModifyDataCmd
		client.ToMsg(scmd.Data, &cmd)
		if len(cmd.Key) == 0 {
			ret = client.CmdRetResult(CodeType_MsgMissParam, nil, "key is nil")
			return
		}
		return client.doQueryData(state, accID, cmd.Key)
	case S_CmdType_Receipt:
		brk = true
	}
	ret = client.CmdRetResult(CodeType_CmdNotFound, nil, "cmd type mismatch")
	return
}

func (client *RemoteAppClient) CmdRetResult(tp CodeType, data []byte, log string) (ret ClientCmd) {
	result := GenPbResult(tp, data, log)
	ret.Type = C_CmdType_CCmdResult
	ret.Data = client.ToBytes(&result)
	return
}

func GenPbResultOK() (result PbResult) {
	return GenPbResult(CodeType_OK, nil, "")
}

func GenPbResult(code CodeType, data []byte, log string) (result PbResult) {
	result.Code = code
	result.Data = data
	result.Log = log
	return
}

func (client *RemoteAppClient) doAddData(state *stateDup, accID string, key string, value []byte) (ret ClientCmd, brk bool) {
	ret.Type = C_CmdType_AddDataRes
	//state.state.SetState(accID, key, value)
	acc := state.state.RemoteAcc(accID)
	var cmd C_AddDataCmd
	if acc == nil {
		result := GenPbResult(CodeType_AccNotFound, nil, "acc not found")
		cmd.Result = &result
		return
	}
	acc.SetKv(key, value)
	result := GenPbResultOK()
	cmd.Result = &result
	ret.Data = client.ToBytes(&cmd)
	return
}

func (client *RemoteAppClient) doDelData(state *stateDup, accID string, key string) (ret ClientCmd, brk bool) {
	ret.Type = C_CmdType_DelDataRes
	if v := state.state.AccGetKv(accID, key); len(v) != 0 {
		var cmd C_DelDataCmd
		result := GenPbResultOK()
		cmd.Result = &result
		cmd.Value = v
		ret.Data = client.ToBytes(&cmd)
		state.state.AccDelKv(accID, key)
		return
	}
	ret = client.CmdRetResult(CodeType_DataNotFound, nil, fmt.Sprintf("key:%v", key))
	return
}

func (client *RemoteAppClient) doModifyData(state *stateDup, accID string, key string, value []byte) (ret ClientCmd, brk bool) {
	ret.Type = C_CmdType_ModifyDataRes
	if v := state.state.AccGetKv(accID, key); len(v) != 0 {
		var cmd C_ModifyDataCmd
		state.state.AccSetKv(accID, key, value)
		result := GenPbResultOK()
		cmd.Result = &result
		ret.Data = client.ToBytes(&cmd)
		return
	}
	ret = client.CmdRetResult(CodeType_DataNotFound, nil, fmt.Sprintf("key:%v", key))
	return
}

func (client *RemoteAppClient) doQueryData(state *stateDup, accID string, key string) (ret ClientCmd, brk bool) {
	ret.Type = C_CmdType_QueryDataRes
	if v := state.state.AccGetKv(accID, key); len(v) != 0 {
		var cmd C_QueryDataCmd
		cmd.Value = v
		result := GenPbResultOK()
		cmd.Result = &result
		ret.Data = client.ToBytes(&cmd)
		return
	}
	ret = client.CmdRetResult(CodeType_DataNotFound, nil, fmt.Sprintf("key:%v", key))
	return
}
