rule_name:= "test rule"
rule_rely_on := map[string]string{
"test":"Test",
}

//--body--

import (
"action"
"fmt"
)

func main() {
resp, err := action.Test("hhhhhh")
fmt.Println("err:", err)
fmt.Println("resp:", resp)
}
