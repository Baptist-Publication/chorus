```
pragma solidity ^0.4.0;

contract Check {

    struct check {
        uint Id;
        uint Amount;
    }
    
    event InputLog (
        uint Id,
        uint Amount
        );
    
    mapping (uint => check) checkInfos;
    
    function createCheckInfos( uint Id, uint Amount) {
        InputLog(Id,Amount);
        check c;
        c.Id = Id;
        c.Amount = Amount;

        checkInfos[Id] = c;
    }
    
    function getPremiumInfos(uint Id) public constant returns(uint) {
        return (
        checkInfos[Id].Amount
        );
    }
}
```



account

```
privkey:  1956be6c7e5128bd1c44c389ba21bd63cfb054d5adb2ab3f8d968f74b9bd0b6b
pubkey: 04e329956f162146cad0d07ddecb5f95329542b1b3badf5b4fe507fd6f0556118326d251e051bdc080bd0f50c0b94d054bcd9b25a2177003296e76a8c07d9b6e22
address:  addafebb1c4618f8f8b452dab6d53721f1d9fda6
```



操作步骤

查询当前nonce

```
./anntool --backend="tcp://0.0.0.0:52001" nonce --address addafebb1c4618f8f8b452dab6d53721f1d9fda6

query result: 9
```

部署合约

```
./anntool --backend="tcp://0.0.0.0:52001" contract create --callf /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample.json --abif /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample.abi --nonce 9

tx result: 0x0a6827ac6ca51a6003b5b4408163091fdc0cfb5c1ddbd98b19b632da9d3e7815
contract address: 0xaae2ea74bf2287f23998b731c97e6c82389d93f2
```



查询合约是否存在

```
./anntool --backend="tcp://0.0.0.0:52001" contract exist --callf /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample_exist.json

Yes!!!
```



查询当前nonce（检查）

```
query result: 10
```



call

```
./anntool --backend="tcp://0.0.0.0:52001" contract execute --callf /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample_execute.json --abif /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample.abi --nonce 10

call data: a6226f2100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000064
```



查询nonce（很重要  万一没成功呢哈哈哈）

```
query result: 11
```



read

```
./anntool --backend="tcp://0.0.0.0:52001" contract read --callf /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample_read.json --abif /Users/wicky/workspace/src/gitlab.zhonganonline.com/ann/test/ann-chain/sample.abi --nonce 11

query result: 0000000000000000000000000000000000000000000000000000000000000064
parse result: *big.Int 100
```

