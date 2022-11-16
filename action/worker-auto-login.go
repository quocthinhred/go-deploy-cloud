package action

import (
	"encoding/json"
	"example.com/micro/client"
	"example.com/micro/client/login"
	"example.com/micro/model"
	"fmt"
	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"log"
	"os"
)

func AutoLoginTask() {
	fmt.Println("Worker running!")
	content, err := os.ReadFile("./account.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var payload []model.Account
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	countFail := 0
	countSuccess := 0
	var listStgFail []string
	var listDevFail []string
	var listNotHaveDomain []string
	for _, v := range payload {
		if v.DomainType == model.STG {
			result := login.FuncLoginStg(client.APIOption{
				Body: payloadToBody(v),
			})
			if result.Status == common.APIStatus.Ok {
				countSuccess++
			} else {
				countFail++
				listStgFail = append(listStgFail, v.Username)
			}
		} else if v.DomainType == model.DEV {
			result := login.FuncLoginDev(client.APIOption{
				Body: payloadToBody(v),
			})
			if result.Status == common.APIStatus.Ok {
				countSuccess++
			} else {
				countFail++
				listDevFail = append(listDevFail, v.Username)
			}
		} else {
			listNotHaveDomain = append(listNotHaveDomain, v.Username)
		}
	}
	fmt.Printf("Worker run task done with %d account!\n%d Successfully\n%d Fail\n%d Missing domainType", len(payload), countSuccess, countFail, len(listNotHaveDomain))

	if countFail > 0 {
		if len(listStgFail) > 0 {
			fmt.Printf("\nList username fail:")
			for _, v := range listStgFail {
				fmt.Printf("\n- %s__%s", "stg", v)
			}
		}

		if len(listDevFail) > 0 {
			for _, v := range listDevFail {
				fmt.Printf("\n- %s__%s", "dev", v)
			}
		}
	}

	if len(listNotHaveDomain) > 0 {
		fmt.Printf("\nList missing domainType:")

		for _, v := range listNotHaveDomain {
			fmt.Printf("\n- %s", v)
		}
	}
	fmt.Println("\nWorker end!")
}

func payloadToBody(account model.Account) map[string]interface{} {
	return map[string]interface{}{
		"username": account.Username,
		"password": account.Password,
		"type":     account.Type,
	}
}
