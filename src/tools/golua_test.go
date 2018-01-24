package tools

import "testing"

const (
	codeOrgA = `                         
	print("in A:"..ent_params["Params1"][1]+ent_params["Params1"][2])  
	for i = 1,  #(ent_params["Params1"]) do             
	        print(i..","..ent_params["Params1"][i])     
	end                                  
	`
	codeOrgB = `                         
	print("in B:"..ent_params["Params2"])               
	`
	globalVarName = "ent_params"

	codeOrgC = `
	ret_params= {                
        	["params1"] = {1,2} ,      
        	["params2"] = "Hello World"
	}                        
	`
)

func TestGoLua(t *testing.T) {
	L := NewLuaState()

	ret, err := ExecLuaWithParam(L, codeOrgC, nil)

	if err != nil {
		t.Error("codeOrgC", err)
	}

	if len(ret) != 2 {
		t.Error("ret_param err:", ret)
	}

	_, err = ExecLuaWithParam(L, codeOrgA, ret)
	if err != nil {
		t.Error("codeOrgA", err, ret)
	}

	_, err = ExecLuaWithParam(L, codeOrgB, ret)
	if err != nil {
		t.Error("codeOrgB", err, ret)
	}
	L.Close()
}
