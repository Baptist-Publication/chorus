package main

func testTx() {
	privkey := defaultPrivKey
	to := ""

	err := sendTx(privkey, to, 1)
	if err != nil {
		return
	}
}
