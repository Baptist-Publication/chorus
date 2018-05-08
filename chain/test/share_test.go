package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Baptist-Publication/chorus/angine/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
	"time"
)

/*
  请将sender address 做为init_token，初始化给予一定balance
*/
var (
	shareReceiver = "0999D61A459023E2A04C3D560708EF4A5C6A5D2A8DCACF61EA3ADE6A8293A5EC"

	nodeprivkey = ""
	nodepub     = ""
)

func init() {
	//get priv_json frile
	privfile := path.Join(runtimePath, "priv_validator.json")
	fmt.Println("privfile path :", privfile)
	if _, err := os.Stat(privfile); err != nil {
		fmt.Println("cannot get priv_validator.json file")
		os.Exit(-1)
	}
	prvJsonBytes, err := ioutil.ReadFile(privfile)
	if err != nil {
		fmt.Println("err :", err)
		os.Exit(-1)
	}
	var privf types.PrivValidatorJSON
	err = json.Unmarshal(prvJsonBytes, &privf)
	if err != nil {
		fmt.Println("error :", err)
		os.Exit(-1)
	}
	nodeprivkey = privf.PrivKey.KeyString()
	nodepub = privf.PubKey.KeyString()
}

//test share transaction
func TestShareTransfer(t *testing.T) {
	msg := make(chan bool)
	nonce, err := getNonce(senderaddress)
	if err != nil {
		t.Error(err)
	}
	go func() {
		args := []string{"share", "send", "--nodeprivkey", nodeprivkey, "--evmprivkey", senderpriv, "--to", shareReceiver, "--value", "888", "--nonce", strconv.FormatUint(nonce, 10)}
		_, err := exec.Command(chorustoolPath, args...).Output()
		if err != nil {
			t.Error(err)
		}
		close(msg)
	}()
	<-msg
	time.Sleep(time.Second * 1)
	args := []string{"query", "share", "--account_pubkey", shareReceiver}
	outs, err := exec.Command(chorustoolPath, args...).Output()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(shareReceiver, " balance :", string(outs))
}

//test share guaranty
func TestGuaranty(t *testing.T) {
	msg := make(chan bool)
	nonce, err := getNonce(senderaddress)
	if err != nil {
		t.Error(err)
	}
	arg := []string{"query", "share", "--account_pubkey", nodepub}
	out1, err := exec.Command(chorustoolPath, arg...).Output()
	if err != nil {
		t.Error(err)
	}
	go func() {
		args := []string{"share", "guarantee", "--nodeprivkey", nodeprivkey, "--evmprivkey", senderpriv, "--value", "123", "--nonce", strconv.FormatUint(nonce, 10)}
		_, err := exec.Command(chorustoolPath, args...).Output()
		if err != nil {
			t.Error(err)
		}
		close(msg)
	}()
	<-msg
	time.Sleep(time.Second * 1)
	out2, err := exec.Command(chorustoolPath, arg...).Output()
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(out1, out2) {
		t.Error("guarantee failed")
	}
}

//test share redeem
func TestRedeem(t *testing.T) {
	msg := make(chan bool)
	nonce, err := getNonce(senderaddress)
	if err != nil {
		t.Error(err)
	}
	arg := []string{"query", "share", "--account_pubkey", nodepub}
	out1, err := exec.Command(chorustoolPath, arg...).Output()
	if err != nil {
		t.Error(err)
	}
	go func() {
		args := []string{"share", "redeem", "--nodeprivkey", nodeprivkey, "--evmprivkey", senderpriv, "--value", "123", "--nonce", strconv.FormatUint(nonce, 10)}
		_, err := exec.Command(chorustoolPath, args...).Output()
		if err != nil {
			t.Error(err)
		}
		close(msg)
	}()
	<-msg
	time.Sleep(time.Second * 1)
	out2, err := exec.Command(chorustoolPath, arg...).Output()
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(out1, out2) {
		t.Error("guarantee failed")
	}
}

func TestClean(t *testing.T) {
	node := <-nodeChan
	node.Process.Kill()
	exec.Command("rm", []string{"./node.*", "./client.*"}...)
}
