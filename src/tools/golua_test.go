package tools

import "testing"

const (
	codeOrgC = `
	function main(params)
		params = {
			["params1"] = {1,2} ,      
			["params2"] = "Hello World"
		}
		return params;
	end
	`
	codeOrgA = `                         
	function main(params)
		for i = 1, #(params["params1"]) do
			print(i..","..params["params1"][i])
		end
		return nil
	end
	`
	codeOrgB = `                         
	function main(params)
		print("in B:"..params["params2"])               
		return nil
	end
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

	L = NewLuaState()
	_, err = ExecLuaWithParam(L, codeOrgA, ret)
	if err != nil {
		t.Error("codeOrgA", err, ret)
	}

	L = NewLuaState()
	_, err = ExecLuaWithParam(L, codeOrgB, ret)
	if err != nil {
		t.Error("codeOrgB", err, ret)
	}
	L.Close()
}
