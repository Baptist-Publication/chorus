
## remote_app

`example_remoteapp.go`是remote-app外部实现逻辑处理的示例，example/sdk/是外部逻辑可以调用的对链访问的接口

执行示例:
	$./build/anntool --backend="tcp://0.0.0.0:50007" --target="app1" tx remote --privkey="ce47f86ea8c6ee8dd9b6568d8c3e826b1294bf7da6293bb58b2eff838728a50f" --payload="add" --nonce=11
	$./build/anntool --backend="tcp://0.0.0.0:50007" --target="app1" tx remote --privkey="ce47f86ea8c6ee8dd9b6568d8c3e826b1294bf7da6293bb58b2eff838728a50f" --payload="query" --nonce=12

查交易:
	$./build/anntool --backend="tcp://0.0.0.0:50007" --target="remoteapp" query rmtreceipt --hash="0xc96b7ba9f5f0056d2d9edaeb679a6aecaf3a956048a487af28101f1a7f7e4f0d"

TODOs:
1. 缺少安全检查，和沙箱环境
2. sdk支持的基本链查询操作待完善
	增加操作需要修改remote/rpc_service.proto


## event code by lua

go和lua之间传递数据(类型定义在`src/types/types.go`):

	go向lua输入参数变量名为"ent_params"
	lua向go输出参数变量名为"ret_params"
	变量类型是map[string]interface{}，值的类型由监听和被监听的组织触发的code逻辑决定

```go
 const (                                      
	codeOrgA = `                         
	print("in A:"..ent_params[1]+ent_params[2])  
	for i = 1,  #(ent_params) do             
	        print(i..","..ent_params[i])     
	end                                  
	`                                    
	codeOrgB = `                         
	print("in B:", ent_params)               
	`                                    
	globalVarName = "ent_params"             

	codeOrgC = `
	ret_params= {                
        	["orgA"] = {1,2} ,      
        	["orgB"] = "Hello World"
	}                        
	`
 )                                 
```
上述例子中，组织A(orgA)和组织B(orgB)在组织C(orgC)上注册了监听事件，codeOrgA、codeOrgB、codeOrgC是注册在各组织的代码，C上的交易触发了A和B的事件，参数格式和内容由组织之间协商。

