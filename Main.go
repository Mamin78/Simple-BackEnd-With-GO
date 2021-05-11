package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
)

var taskDoneMSG = "Task done successfully!"
var historySent = "History sent successfully!"
var notFoundErr = "Not found error!"
var badMethodErr = "Bad method error!"
var badTaskErr = "Bad task error!"
var extraFieldErr = "Extra field error!"

type Request struct {
	Task    string                 `json:"task"`
	Numbers []int                  `json:"numbers"`
	X       map[string]interface{} `json:"-"`
}

type FloatRequest struct {
	Task    string                 `json:"task"`
	Numbers []interface{}          `json:"numbers"`
	X       map[string]interface{} `json:"-"`
}

type MeanResponse struct {
	Task    string  `json:"task"`
	Numbers []int   `json:"numbers"`
	Answer  float64 `json:"answer"`
	Code    int     `json:"code"`
	Message string  `json:"message"`
}

func NewMeanResponse(task, message string, numbers []int, answer float64, code int) *MeanResponse {
	mr := new(MeanResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Answer = answer
	mr.Code = code
	mr.Message = message
	return mr
}

type SortResponse struct {
	Task    string `json:"task"`
	Numbers []int  `json:"numbers"`
	Answer  []int  `json:"answer"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewSortResponse(task, message string, numbers, answer []int, code int) *SortResponse {
	mr := new(SortResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Answer = answer
	mr.Code = code
	mr.Message = message
	return mr
}

type BadResponse struct {
	Task    string `json:"task"`
	Numbers []int  `json:"numbers"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewBadResponse(task, message string, numbers []int, code int) *BadResponse {
	mr := new(BadResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Code = code
	mr.Message = message
	return mr
}

type ExtraFieldResponse struct {
	Task    string   `json:"task"`
	Numbers []int    `json:"numbers"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Fields  []string `json:"fields"`
}

func NewExtraFieldResponse(task, message string, numbers []int, code int, fields []string) *ExtraFieldResponse {
	mr := new(ExtraFieldResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Code = code
	mr.Message = message
	mr.Fields = fields
	return mr
}

type Response interface{}

type SavedMeanResponse struct {
	Task    string  `json:"task"`
	Numbers []int   `json:"numbers"`
	Answer  float64 `json:"answer"`
}

func NewSavedMeanResponse(task string, numbers []int, answer float64) *SavedMeanResponse {
	mr := new(SavedMeanResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Answer = answer
	return mr
}

type SavedSortResponse struct {
	Task    string `json:"task"`
	Numbers []int  `json:"numbers"`
	Answer  []int  `json:"answer"`
}

func NewSavedSortResponse(task string, numbers, answer []int) *SavedSortResponse {
	mr := new(SavedSortResponse)
	mr.Task = task
	mr.Numbers = numbers
	mr.Answer = answer
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
	Size    int        `json:"size"`
	History []Response `json:"history"`
	Code    int        `json:"code"`
	Message string     `json:"message"`
}

var serverHistory Server

func getRequestBody(req *http.Request) string {
	reqBody, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		println(err)
	}
	return string(reqBody)
}
func unmarshalRequest(w http.ResponseWriter, reqBody string) *Request {
	var request *Request
	err := json.Unmarshal([]byte(reqBody), &request)
	if err != nil {
		//panic("Unmarshall error")
		temp := unmarshalFloatRequest(w, reqBody)
		if reflect.TypeOf(temp.Numbers) == reflect.TypeOf(reflect.Interface) {

			fmt.Fprintln(w, reflect.TypeOf(temp.Numbers[0]))
			if reflect.TypeOf(temp.Numbers[0]) != reflect.TypeOf(reflect.Int) {
				fmt.Fprintln(w, "Type of numbers is unacceptable!")
			}
		} else {
			request = nil
		}

	}
	return request
}

func unmarshalFloatRequest(w http.ResponseWriter, reqBody string) *FloatRequest {
	var request *FloatRequest
	err := json.Unmarshal([]byte(reqBody), &request)
	if err != nil {
		fmt.Fprintln(w, "Type of numbers is unacceptable2222!")
		//panic("Unmarshall error")
	}
	return request
}

func CalcMean(numbers []int) float64 {
	sum := 0
	for i := 0; i < len(numbers); i++ {
		sum += numbers[i]
	}
	mean := float64(sum) / float64(len(numbers))
	return mean
}

func writePostMethodError(w http.ResponseWriter, req *http.Request, code int, message string) {
	reqBody := getRequestBody(req)
	request := unmarshalRequest(w, reqBody)
	resp := NewBadResponse(request.Task, message, request.Numbers, code)
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
		resp := NewExtraFieldResponse(request.Task, extraFieldErr, request.Numbers, http.StatusNotFound, listOfKeys(request.X))
		b, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(b))
		return
	}

	var savedResponse Response
	var b []byte
	if request.Task == "mean" {
		mean := CalcMean(request.Numbers)
		resp := NewMeanResponse(request.Task, taskDoneMSG, request.Numbers, mean, 200)
		savedResponse = NewSavedMeanResponse(request.Task, request.Numbers, mean)
		serverHistory.History = append(serverHistory.History, savedResponse)
		b, _ = json.Marshal(resp)
		fmt.Fprintln(w, string(b))
	} else if request.Task == "sort" {
		answer := make([]int, len(request.Numbers))
		copy(answer, request.Numbers)
		sort.Ints(answer)
		resp := NewSortResponse(request.Task, taskDoneMSG, request.Numbers, answer, 200)
		savedResponse = NewSavedSortResponse(request.Task, request.Numbers, answer)
		serverHistory.History = append(serverHistory.History, savedResponse)
		b, _ = json.Marshal(resp)
		fmt.Fprintln(w, string(b))
	} else {
		resp := NewBadResponse(request.Task, badTaskErr, request.Numbers, http.StatusNotFound)
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
