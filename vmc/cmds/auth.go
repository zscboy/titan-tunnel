package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"titan-vm/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

const (
	vmapiMultipass           = "multipass"
	vmapiLibvirt             = "libvirt"
	linuxSSHAuthFile         = "/root/.ssh/authorized_keys"
	linuxMultipassAuthFile   = "/var/snap/multipass/common/data/multipassd/authenticated-certs/multipass_client_certs.pem"
	windowsMultipassAuthFile = "C://ProgramData//Multipass//data//authenticated-certs/multipass_client_certs.pem"
)

type AuthSSHAndMultipass struct {
	vmapi string
	// dm *downloader.Manager
}

func NewAuthSSHAndMultipass(vmapi string) *AuthSSHAndMultipass {
	return &AuthSSHAndMultipass{}
}

func (auth *AuthSSHAndMultipass) Auth(req []byte) *pb.CmdAuthSSHAndMultipassResponse {
	logx.Debug("Auth")
	return auth.auth(req)
}

func (auth *AuthSSHAndMultipass) auth(req []byte) *pb.CmdAuthSSHAndMultipassResponse {
	authRequest := &pb.CmdAuthSSHAndMultipassRequest{}
	err := proto.Unmarshal(req, authRequest)
	if err != nil {
		return &pb.CmdAuthSSHAndMultipassResponse{ErrMsg: err.Error()}
	}

	if runtime.GOOS == "linux" {
		if err := auth.addSSHPubKey(authRequest.SshPubKey); err != nil {
			return &pb.CmdAuthSSHAndMultipassResponse{ErrMsg: err.Error()}
		}
	}

	if auth.vmapi == "multipass" {
		if err := auth.addMultipassCert(authRequest.MultipassCert); err != nil {
			return &pb.CmdAuthSSHAndMultipassResponse{ErrMsg: err.Error()}
		}
	}

	return &pb.CmdAuthSSHAndMultipassResponse{Success: true}

}

func (auth *AuthSSHAndMultipass) addSSHPubKey(pubKey []byte) error {
	if err := auth.createFileIfNotExist(linuxSSHAuthFile); err != nil {
		return err
	}

	authFile, err := os.ReadFile(linuxSSHAuthFile)
	if err != nil {
		return err
	}

	authFileString := string(authFile)
	if strings.Contains(authFileString, string(pubKey)) {
		return nil
	}

	file, err := os.OpenFile(linuxSSHAuthFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(pubKey); err != nil {
		return err
	}
	return nil
}

func (auth *AuthSSHAndMultipass) createFileIfNotExist(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func (auth *AuthSSHAndMultipass) addMultipassCert(cert []byte) error {
	var filePath string
	if runtime.GOOS == "linux" {
		filePath = linuxMultipassAuthFile
	} else if runtime.GOOS == "windows" {
		filePath = windowsMultipassAuthFile
	} else {
		return fmt.Errorf("unsupport os %s", runtime.GOOS)
	}

	authFile, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	authFileString := string(authFile)
	if strings.Contains(authFileString, string(cert)) {
		return nil
	}

	file, err := os.OpenFile(linuxSSHAuthFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(cert); err != nil {
		return err
	}
	return nil
}
