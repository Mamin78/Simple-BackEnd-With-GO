package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
)

var taskDoneMSG = "Task done successfully!"
var historySent = "History sent successfully!"
var notFoundErr = "Not found error!"
var badMethodErr = "Bad method error!"
var badTaskErr = "Bad task error!"
var extraFieldErr = "Extra field error!"
var typeErr = "Type mismatch!"
var incorrectErr = "Your input is incorrect"

type Request struct {
	Task interface{}            `json:"task"`
	X    map[string]interface{} `json:"-"`
}

type SliceRequest struct {
	Request
	Numbers []interface{} `json:"numbers"`
}

type IntRequest struct {
	Numbers []int `json:"numbers"`
}

type NotSliceRequest struct {
	Request
	Numbers interface{} `json:"numbers"`
}

type BaseResponse struct {
	Task    string        `json:"task"`
	Numbers []interface{} `json:"numbers"`
}

func NewBaseResponse(task string, numbers []interface{}) *BaseResponse {
	mr := new(BaseResponse)
	mr.Task = task
	mr.Numbers = numbers
	return mr
}

type SavedResponse struct {
	BaseResponse
	Answer interface{} `json:"answer"`
}

func NewSavedResponse(task string, numbers []interface{}, answer interface{}) *SavedResponse {
	mr := new(SavedResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Answer = answer
	return mr
}

type Response struct {
	SavedResponse
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewResponse(task, message string, numbers []interface{}, answer interface{}, code int) *Response {
	mr := new(Response)
	mr.Task = task
	mr.Numbers = numbers
	mr.Code = code
	mr.Message = message
	mr.Answer = answer
	return mr
}

type BadResponse struct {
	BaseResponse
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewBadResponse(task, message string, numbers []interface{}, code int) *BadResponse {
	mr := new(BadResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Code = code
	mr.Message = message
	return mr
}

type ExtraFieldResponse struct {
	BadResponse
	Fields []string `json:"fields"`
}

func NewExtraFieldResponse(task, message string, numbers []interface{}, code int, fields []string) *ExtraFieldResponse {
	mr := new(ExtraFieldResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Code = code
	mr.Message = message
	mr.Fields = fields
	return mr
}

type MethodError struct {
	Method  string `json:"method"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewMethodError(method, message string, code int) *MethodError {
	mr := new(MethodError)
	mr.Method = method
	mr.Code = code
	mr.Message = message
	return mr
}

type Server struct {
	Size    int           `json:"size"`
	History []interface{} `json:"history"`
	Code    int           `json:"code"`
	Message string        `json:"message"`
}

var serverHistory Server

func getStringValue(x interface{}) string {
	var result = ""
	switch x.(type) {
	case string:
		return fmt.Sprintf("%v", x)
	default:
		return result
	}
}

func getRequestBody(req *http.Request) string {
	reqBody, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		println(err)
	}
	return string(reqBody)
}

func unmarshalRequest(w http.ResponseWriter, reqBody string) *SliceRequest {
	var request *SliceRequest
	err := json.Unmarshal([]byte(reqBody), &request)
	if err != nil {
		notSlice := unmarshalNotSliceRequest(w, reqBody)
		if notSlice == nil {
			return nil
		}
		temp := make([]interface{}, 1)
		temp[0] = notSlice.Numbers
		resp := NewBadResponse(getStringValue(notSlice.Task), typeErr, temp, http.StatusNotFound)
		b, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(b))
		return nil
	}
	var intRequest *IntRequest
	err = json.Unmarshal([]byte(reqBody), &intRequest)
	if err != nil {
		resp := NewBadResponse(fmt.Sprintf("%v", request.Task), typeErr, request.Numbers, http.StatusNotFound)
		b, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(b))
		return nil
	}
	return request

}

//
func unmarshalNotSliceRequest(w http.ResponseWriter, reqBody string) *NotSliceRequest {
	var request *NotSliceRequest
	err := json.Unmarshal([]byte(reqBody), &request)
	if err != nil {
		fmt.Fprintln(w, incorrectErr)
		return nil
	}
	return request
}

func CalcMean(numbers []interface{}) float64 {
	sum := 0
	for i := 0; i < len(numbers); i++ {
		val := int(numbers[i].(float64))
		sum += val
	}
	return float64(sum) / float64(len(numbers))
}

func writePostMethodError(w http.ResponseWriter, req *http.Request, code int, message string) {
	reqBody := getRequestBody(req)
	request := unmarshalRequest(w, reqBody)
	resp := NewBadResponse(getStringValue(request.Task), message, request.Numbers, code)
	b, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(b))
}

func writeMethodError(w http.ResponseWriter, req *http.Request, code int, method, message string) {
	err := NewMethodError(method, message, code)
	b, _ := json.Marshal(err)
	fmt.Fprintln(w, string(b))
}

func listOfKeys(X map[string]interface{}) []string {
	keys := make([]string, 0, len(X))
	for k := range X {
		keys = append(keys, k)
	}
	return keys
}

func interfaceToInt(answer []interface{}) []int {
	res := make([]int, len(answer))
	for i := 0; i < len(answer); i++ {
		iAreaId := int(answer[i].(float64))
		res[i] = iAreaId
	}
	return res
}

func calculator(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		writeMethodError(w, req, http.StatusNotFound, req.Method, badMethodErr)
		return
	}
	reqBody := getRequestBody(req)
	println(reqBody)
	request := unmarshalRequest(w, reqBody)
	if request == nil {
		return
	}
	if err := json.Unmarshal([]byte(reqBody), &request.X); err != nil {
		panic(err)
	}
	delete(request.X, "task")
	delete(request.X, "numbers")

	if len(request.X) > 0 {
		resp := NewExtraFieldResponse(getStringValue(request.Task), extraFieldErr, request.Numbers, http.StatusNotFound, listOfKeys(request.X))
		b, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(b))
		return
	}

	var savedResponse *SavedResponse
	var b []byte
	if request.Task == "mean" {
		mean := CalcMean(request.Numbers)
		resp := NewResponse(getStringValue(request.Task), taskDoneMSG, request.Numbers, mean, 200)
		savedResponse = NewSavedResponse(getStringValue(request.Task), request.Numbers, mean)
		serverHistory.History = append(serverHistory.History, savedResponse)
		b, _ = json.Marshal(resp)
		fmt.Fprintln(w, string(b))
	} else if request.Task == "sort" {
		answerInterface := make([]interface{}, len(request.Numbers))
		copy(answerInterface, request.Numbers)
		answer := interfaceToInt(answerInterface)
		sort.Ints(answer)
		resp := NewResponse(getStringValue(request.Task), taskDoneMSG, request.Numbers, answer, 200)
		savedResponse = NewSavedResponse(getStringValue(request.Task), request.Numbers, answer)
		serverHistory.History = append(serverHistory.History, savedResponse)
		b, _ = json.Marshal(resp)
		fmt.Fprintln(w, string(b))
	} else {
		message := typeErr
		switch request.Task.(type) {
		case string:
			message = badTaskErr
		}
		resp := NewBadResponse(fmt.Sprintf("%v", request.Task), message, request.Numbers, http.StatusNotFound)
		b, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(b))
	}

	serverHistory.Size = len(serverHistory.History)
}

func history(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		writeMethodError(w, req, http.StatusNotFound, req.Method, badMethodErr)
		return
	}
	b, _ := json.Marshal(serverHistory)
	fmt.Fprintln(w, string(b))
}

func badEndPoint(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		writePostMethodError(w, req, 404, notFoundErr)
	} else {
		writeMethodError(w, req, http.StatusNotFound, req.Method, notFoundErr)
	}
}

func main() {
	serverHistory.Code = 200
	serverHistory.Message = historySent

	http.HandleFunc("/", badEndPoint)
	http.HandleFunc("/calculator", calculator)
	http.HandleFunc("/history", history)

	http.ListenAndServe(":8080", nil)
}
