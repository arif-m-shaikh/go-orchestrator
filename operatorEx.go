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
	BodyTransSpec     string
	BodyTransFunc     string
	ResponseSchema    string
	EnvStoreTransSpec string
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

	//-------------------------| Replace in Body
	body := apiContext.Body
	if body != nil {
		if apiContext.EnvVarInBody {
			body, err = replaceVarInJson(body)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		}

		if apiContext.BodyTransSpec != "" {
			body, err = transformData(apiContext.BodyTransSpec, body)
			if err != nil {
				log.Fatal(err)
				return nil, err
			}
		} else if apiContext.BodyTransFunc != "" {
			body, err = transformDataUsingCode(apiContext.Name,
				apiContext.BodyTransFunc,
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

	if apiContext.EnvStoreTransSpec != "" {
		err := storeVarInJson(apiContext, resData)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	return resData, nil
}
