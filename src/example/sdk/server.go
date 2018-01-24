package sdk

import (
	"fmt"
	"net"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	ethcmn "github.com/Baptist-Publication/chorus-module/lib/eth/common"
	remote "github.com/Baptist-Publication/chorus/src/chain/app/remote"
)

type ServerApp interface {
	CheckTx(account ethcmn.Address, txBytes []byte) error
	Execute(helper *Helper, blockHeight int64, account ethcmn.Address, txBytes []byte) (receiptData []byte, err error)
}

func errToPbResult(err error, code remote.CodeType) (rs remote.PbResult) {
	if err != nil {
		rs.Log = err.Error()
		rs.Code = code
	}
	return
}

type Server struct {
	app ServerApp
}

func NewServer(sapp ServerApp) *Server {
	return &Server{
		app: sapp,
	}
}

func (s *Server) CheckTx(ctx context.Context, req *remote.ClientCheckCmd) (*remote.ServerCheckRes, error) {
	acc := ethcmn.HexToAddress(req.AccountID)
	err := s.app.CheckTx(acc, req.Tx)
	res := errToPbResult(err, remote.CodeType_ExternalErr)
	sCheck := &remote.ServerCheckRes{
		Result: &res,
	}
	return sCheck, nil
}

func (s *Server) Execute(do remote.RemoteAppService_ExecuteServer) error {
	var revErr error
	var cmdTx *remote.ClientCmd
	cmdTx, revErr = do.Recv()
	if revErr != nil {
		fmt.Println("recv error:", revErr)
		return revErr
	}
	if cmdTx.Type != remote.C_CmdType_Tx {
		err := fmt.Errorf("cmd type error:", cmdTx.Type)
		fmt.Println(err)
		return err
	}
	var cmdTxData remote.C_TxCmd
	if err := ToMsg(cmdTx.Data, &cmdTxData); err != nil {
		fmt.Println("to msg error:", err)
		return err
	}
	var h Helper
	h.Init(do)
	receipt, err := s.app.Execute(&h, cmdTxData.BlockHeight, ethcmn.HexToAddress(cmdTx.AccountID), cmdTxData.Tx)
	if err != nil {
		return err
	}
	ret := GenCmd(nil, remote.S_CmdType_Receipt, receipt)
	return do.Send(&ret)

}

func (s *Server) Start(ip string) error {
	if s.app == nil {
		return fmt.Errorf("app is not regitered")
	}
	lis, err := net.Listen("tcp", ip)
	if err != nil {
		return fmt.Errorf("fail to listen,err:", err)
	}
	server := grpc.NewServer()
	remote.RegisterRemoteAppServiceServer(server, s)
	return server.Serve(lis)
}
