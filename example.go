package goflow

import (
	"errors"
	"net/http"
)

func analyticsJob() *Job {
	j := &Job{
		Name:     "exampleAnalytics",
		Schedule: "* * * * *",
	}

	var logDataContext ApiContext
	logDataContext.Init()
	logDataContext.Name = "LogData"
	logDataContext.HttpMethod = http.MethodGet
	logDataContext.ApiURL = "http://localhost:9090/api/logdata"

	//=============================================| Get the temparature data
	specTemp := `[{
		"operation": "shift",
		"over": "$",
		"spec": {
			"temparature": "name",
			"reading": "value"
		}
		}]`

	var getTemperatureContext ApiContext
	getTemperatureContext.Init()
	getTemperatureContext.Name = "GetTemparature"
	getTemperatureContext.HttpMethod = http.MethodGet
	getTemperatureContext.ApiURL = "http://localhost:9090/api/temperature"
	getTemperatureContext.ResponseTransSpec = specTemp

	j.Add(&Task{
		Name:     "GetTemparature",
		Operator: getTemperatureContext,
	})

	j.Add(&Task{
		Name:     "LogData1",
		Operator: logDataContext,
	})

	//=============================================| Get the humidity data
	specHumidity := `[{
		"operation": "shift",
		"over": "$",
		"spec": {
			"humidity": "name",
			"reading": "value"
		}
		}]`

	var getHumidityContext ApiContext
	getHumidityContext.Init()
	getHumidityContext.Name = "GetHumidity"
	getHumidityContext.HttpMethod = http.MethodGet
	getHumidityContext.ApiURL = "http://localhost:9090/api/humidity"
	getHumidityContext.ResponseTransSpec = specHumidity

	j.Add(&Task{
		Name:     "GetHumidity",
		Operator: getHumidityContext,
	})

	j.Add(&Task{
		Name:     "LogData2",
		Operator: logDataContext,
	})
	//=============================================| Process the data
	var processHVACContext ApiContext
	processHVACContext.Init()
	processHVACContext.Name = "ProcessHVAC"
	processHVACContext.HttpMethod = http.MethodPost
	processHVACContext.ApiURL = "http://localhost:9090/api/processhvac"

	j.Add(&Task{
		Name:        "ProcessHVAC",
		Operator:    processHVACContext,
		TriggerRule: "allSuccessful",
	})

	j.Add(&Task{
		Name:     "LogData3",
		Operator: logDataContext,
	})
	//=============================================| Process the data
	var storeHVACContext ApiContext
	storeHVACContext.Init()
	storeHVACContext.Name = "StoreHVAC"
	storeHVACContext.HttpMethod = http.MethodPost
	storeHVACContext.ApiURL = "http://localhost:9090/api/storehvac"

	j.Add(&Task{
		Name:        "StoreHVAC",
		Operator:    storeHVACContext,
		TriggerRule: "allDone",
	})
	j.SetDownstream(j.Task("GetTemparature"), j.Task("LogData1"))
	j.SetDownstream(j.Task("LogData1"), j.Task("ProcessHVAC"))
	j.SetDownstream(j.Task("GetHumidity"), j.Task("LogData2"))
	j.SetDownstream(j.Task("LogData2"), j.Task("ProcessHVAC"))
	j.SetDownstream(j.Task("ProcessHVAC"), j.Task("LogData3"))
	j.SetDownstream(j.Task("LogData3"), j.Task("StoreHVAC"))
	return j
}

// Crunch some numbers
func complexAnalyticsJob() *Job {
	j := &Job{
		Name:     "exampleComplexAnalytics",
		Schedule: "* * * * *",
	}

	j.Add(&Task{
		Name:     "sleepOne",
		Operator: Command{Cmd: "sleep", Args: []string{"1"}},
	})
	j.Add(&Task{
		Name:     "addOneOne",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((1 + 1))"}},
	})
	j.Add(&Task{
		Name:     "sleepTwo",
		Operator: Command{Cmd: "sleep", Args: []string{"2"}},
	})
	j.Add(&Task{
		Name:     "addTwoFour",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}},
	})
	j.Add(&Task{
		Name:     "addThreeFour",
		Operator: Command{Cmd: "sh", Args: []string{"-c", "echo $((3 + 4))"}},
	})
	j.Add(&Task{
		Name:       "whoopsWithConstantDelay",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    5,
		RetryDelay: ConstantDelay{Period: 1},
	})
	j.Add(&Task{
		Name:       "whoopsWithExponentialBackoff",
		Operator:   Command{Cmd: "whoops", Args: []string{}},
		Retries:    1,
		RetryDelay: ExponentialBackoff{},
	})
	j.Add(&Task{
		Name:        "totallySkippable",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'everything succeeded'"}},
		TriggerRule: "allSuccessful",
	})
	j.Add(&Task{
		Name:        "cleanUp",
		Operator:    Command{Cmd: "sh", Args: []string{"-c", "echo 'cleaning up now'"}},
		TriggerRule: "allDone",
	})

	j.SetDownstream(j.Task("sleepOne"), j.Task("addOneOne"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("sleepTwo"))
	j.SetDownstream(j.Task("sleepTwo"), j.Task("addTwoFour"))
	j.SetDownstream(j.Task("addOneOne"), j.Task("addThreeFour"))
	j.SetDownstream(j.Task("sleepOne"), j.Task("whoopsWithConstantDelay"))
	j.SetDownstream(j.Task("sleepOne"), j.Task("whoopsWithExponentialBackoff"))
	j.SetDownstream(j.Task("whoopsWithConstantDelay"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("whoopsWithExponentialBackoff"), j.Task("totallySkippable"))
	j.SetDownstream(j.Task("totallySkippable"), j.Task("cleanUp"))

	return j
}

// PositiveAddition adds two nonnegative numbers. This is just a contrived example to
// demonstrate the usage of custom operators.
type PositiveAddition struct{ a, b int }

// Run implements the custom operation.
func (o PositiveAddition) Run() ([]byte, error) {
	if o.a < 0 || o.b < 0 {
		return []byte{0}, errors.New("Can't add negative numbers")
	}
	var result interface{}
	result = o.a + o.b
	return result.([]byte), nil
}

// Use our custom operation in a job.
func customOperatorJob() *Job {
	j := &Job{Name: "exampleCustomOperator", Schedule: "* * * * *", Active: true}
	j.Add(&Task{Name: "posAdd", Operator: PositiveAddition{5, 6}})
	return j
}
