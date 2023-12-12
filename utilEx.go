package goflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"plugin"
	"regexp"
	"strings"

	"github.com/qntfy/kazaam"
)

var environment map[string]string = make(map[string]string)

func isTokenExist(slice []string, token string) int {
	for i, element := range slice {
		if element == token {
			return i
		}
	}
	return -1
}

func findTokens(str string) []string {
	m := regexp.MustCompile(`{{[A-z0-9_]+}}`)
	tokenList := m.FindAllString(str, -1)
	var uniqueTokenList []string
	for _, eachToken := range tokenList {
		if isTokenExist(uniqueTokenList, eachToken) == -1 {
			uniqueTokenList = append(uniqueTokenList, eachToken)
		}
	}

	return uniqueTokenList
}

func replaceVarInString(inStr string) (string, error) {
	if strings.Contains(inStr, "{{") {
		envVarList := findTokens(inStr)
		for _, eachEnvVar := range envVarList {
			eachTrimVar := strings.Trim(eachEnvVar, "{{")
			eachTrimVar = strings.Trim(eachTrimVar, "}}")
			replaceVar, exists := environment[eachTrimVar]
			if exists {
				strings.ReplaceAll(inStr, eachEnvVar, replaceVar)
			} else {
				err := errors.New("Key not found in environment")
				return "", err
			}
		}
	}

	return inStr, nil
}

func replaceVarInJson(inJson []byte) ([]byte, error) {
	strJson := string(inJson)
	if strings.Contains(strJson, "{{") {
		envVarList := findTokens(strJson)
		for _, eachEnvVar := range envVarList {
			eachTrimVar := strings.Trim(eachEnvVar, "{{")
			eachTrimVar = strings.Trim(eachTrimVar, "}}")
			replaceVar, exists := environment[eachTrimVar]
			if exists {
				strings.ReplaceAll(strJson, eachEnvVar, replaceVar)
			} else {
				err := errors.New("Key not found in environment")
				return nil, err
			}
		}
	}

	return []byte(strJson), nil
}

func storeVarInJson(apiContext ApiContext, resData []byte) error {
	k, err := kazaam.NewKazaam(apiContext.EnvStoreTransSpec)
	if err != nil {
		log.Fatal("Unable to load Kazaam specification file: " + err.Error())
		return err
	}

	envJson, transformError := k.TransformJSONStringToString(string(resData))
	if transformError != nil {
		log.Fatal("Unable to transform message", transformError)
		return err
	}

	var respDataMap map[string]interface{}
	err = json.Unmarshal([]byte(envJson), &respDataMap)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for k, v := range respDataMap {
		environment[k] = v.(string)
	}
	return nil
}

func transformData(spec string, data []byte) ([]byte, error) {
	k, err := kazaam.NewKazaam(spec)
	if err != nil {
		log.Fatal("Unable to load Kazaam specification: " + err.Error())
		return nil, err
	}

	transData, transformError := k.TransformJSONStringToString(string(data))
	if transformError != nil {
		log.Fatal("Unable to transform message", transformError)
		return nil, err
	}

	return []byte(transData), nil
}

func httpCall(httpMethod,
	apiURL string,
	inHeader []byte,
	inBody []byte) ([]byte, error) {

	//====================| Create request
	req, err := http.NewRequest(httpMethod, apiURL, bytes.NewReader(inBody))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//====================| Set the header
	req.Header.Set("Content-Type", "application/json")
	if inHeader != nil {
		var mapHeader map[string]string
		err = json.Unmarshal(inHeader, &mapHeader)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		for k, v := range mapHeader {
			req.Header.Set(k, v)
		}
		//fmt.Println(req.Header)
	}

	//====================| Make the API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	//====================| Convert the return value to json and return
	resData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return resData, nil
}

func createCodeFile(codeName, codeStr string) error {
	file, err := os.Create(codeName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(codeStr)
	if err != nil {
		return err
	}

	return nil
}

func transformDataUsingCode(codeName, codeStr string, inJson []byte) ([]byte, error) {
	codeFile := codeName + ".go"
	compileFile := codeName + ".so"

	//===================| Code file exist? => delete it
	_, err := os.Stat(codeFile)
	if err == nil {
		err := os.Remove(codeFile) //remove the file
		if err != nil {
			log.Fatal("Failed to remove code file- " + codeFile + ": " + err.Error())
			return nil, err
		}
	}

	//===================| Create the code file
	err = createCodeFile(codeFile, codeStr)
	if err != nil {
		log.Fatal("Failed to create code file- " + codeFile + ": " + err.Error())
		return nil, err
	}

	//===================| Build the plugin
	exec.Command("go", "build", "-buildmode=plugin", "-o", compileFile, codeFile).Output()
	plug, err := plugin.Open(compileFile)
	if err != nil {
		log.Fatal("Failed to load plugin- " + compileFile + ": " + err.Error())
		return nil, err
	}

	//===================| Load the plugin
	functionCode, err := plug.Lookup(codeName)
	if err != nil {
		log.Fatal("Failed to load function- " + compileFile + ": " + err.Error())
		return nil, err
	}

	//===================| Load the function
	functionLoaded := functionCode.(func([]byte) ([]byte, error))

	//===================| Call the function
	//fmt.Println(string(inJson))
	outJson, err := functionLoaded(inJson)
	if err != nil {
		log.Fatal("Failed to load function- " + compileFile + ": " + err.Error())
		return nil, err
	}

	//fmt.Println(string(outJson))
	return outJson, nil
}
