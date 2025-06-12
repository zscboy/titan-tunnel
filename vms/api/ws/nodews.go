package ws

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"
	"titan-vm/vms/api/internal/types"
	"titan-vm/vms/model"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	timeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

type NodeWS struct {
	tunMgr *TunnelManager
}

func NewNodeWS(tunMgr *TunnelManager) *NodeWS {
	return &NodeWS{tunMgr: tunMgr}
}

func (ws *NodeWS) ServeWS(w http.ResponseWriter, r *http.Request, req *types.NodeWSRequest) error {
	logx.Infof("NodeWS.ServeWS %s, %v", r.URL.Path, req)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err
	}

	node, err := model.GetNode(ws.tunMgr.redis, req.NodeId)
	if err != nil {
		logx.Errorf("ServeWS, get node %s", err.Error())
		return err
	}

	if node == nil {
		node = &model.Node{Id: req.NodeId, PubKey: req.Pubkey, RegisterAt: time.Now().Format(timeLayout)}
	}

	node.OS = req.OS
	node.VmAPI = req.VMAPI
	node.IP = ip
	node.Online = true
	node.LoginAt = time.Now().Format(timeLayout)

	if err := ws.verifySign(node.PubKey, req.Sign, node.Id); err != nil {
		return err
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ws.tunMgr.acceptWebsocket(c, node)

	return nil
}

func (ws *NodeWS) verifySign(pubKey, sign, nodeId string) error {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return err
	}

	signBytes, err := hex.DecodeString(sign)
	if err != nil {
		return err
	}
	return ws.verifySignatureWithSecp256k1PubKey(pubKeyBytes, []byte(nodeId), signBytes)
}

func (ws *NodeWS) verifySignatureWithSecp256k1PubKey(pubKey []byte, msg []byte, signatrue []byte) error {
	var newPubKey = secp256k1.PubKey(pubKey)
	if !newPubKey.VerifySignature(msg, signatrue) {
		return fmt.Errorf("verifySignatureWithSecp256k1PubKey, VerifySignature failed")
	}

	return nil
}
