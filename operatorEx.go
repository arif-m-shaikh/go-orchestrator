package goflow

import (
	"log"
)

type ApiContext struct {
	Name              string
	HttpMethod        string
	EnvVarInURL       bool
	ApiURL            string
	EnvVarInHeader    bool
	Header            []byte
	EnvVarInBody      bool
	Body              []byte
	RequestSchema     string
	RequestTransSpec  string
	RequestTransFunc  string
	ResponseSchema    string
	ResponseTransSpec string
	ResponseTransFunc string
	EnvStoreTransSpec string
}

func (apiContext *ApiContext) Init() {
	apiContext.Name = ""
	apiContext.HttpMethod = ""
	apiContext.EnvVarInURL = false
	apiContext.ApiURL = ""
	apiContext.EnvVarInHeader = false
	apiContext.Header = nil
	apiContext.EnvVarInBody = false
	apiContext.Body = nil
	apiContext.RequestSchema = ""
	apiContext.RequestTransSpec = ""
	apiContext.RequestTransFunc = ""
	apiContext.ResponseSchema = ""
	apiContext.ResponseTransSpec = ""
	apiContext.ResponseTransFunc = ""
	apiContext.EnvStoreTransSpec = ""
}

func (apiContext ApiContext) Run() ([]byte, error) {
	//-------------------------| Replace in URL
	apiUrl := apiContext.ApiURL
	var err error
	if apiContext.EnvVarInURL {
		apiUrl, err = replaceVarInString(apiContext.ApiURL)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	//-------------------------| Replace in Header
	header := apiContext.Header
	if header != nil {
		if apiContext.EnvVarInHeader {
			header, err = replaceVarInJson(header)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}
	}

	//-------------------------| Replace in Request
	body := apiContext.Body
	if body != nil {
		if apiContext.EnvVarInBody {
			body, err = replaceVarInJson(body)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}

		if apiContext.RequestTransSpec != "" {
			body, err = transformData(apiContext.RequestTransSpec, body)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		} else if apiContext.RequestTransFunc != "" {
			body, err = transformDataUsingCode(apiContext.Name,
				apiContext.RequestTransFunc,
				body)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}
	}

	//-------------------------|
	resData, err := httpCall(apiContext.HttpMethod,
		apiUrl,
		header,
		body)
	if err != nil {
		log.Fatal(err)
	}

	//-------------------------| Replace in Response
	if resData != nil {
		if apiContext.EnvStoreTransSpec != "" {
			err := storeVarInJson(apiContext, resData)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}

		if apiContext.ResponseTransSpec != "" {
			resData, err = transformData(apiContext.ResponseTransSpec, resData)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		} else if apiContext.ResponseTransFunc != "" {
			resData, err = transformDataUsingCode(apiContext.Name,
				apiContext.ResponseTransFunc,
				resData)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}
	}

	return resData, nil
}
