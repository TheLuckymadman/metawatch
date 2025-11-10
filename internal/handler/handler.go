package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	
	"github.com/TheLuckymadman/metawatch/internal/model"
)

func MetricSetterHandler(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		// if r.Header.Get("Content-Type") != "text/plain" {
		// 	http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
		// 	w.WriteHeader(http.StatusMethodNotAllowed)
		// 	return
		// }
		
		metricType := chi.URLParam(r, "type")
		restPath := chi.URLParam(r, "*")
		reqPath := regexp.MustCompile(`^(.+)/(.+)$`)
		matches := reqPath.FindStringSubmatch(restPath)
		if len(matches) < 3 {
			http.Error(w, "invalid path format . Use: /update/metrictype/metricname/values", http.StatusNotFound)
			return
		}
		metricName := matches[1]
		metricValue := matches[2]
		agentIP := strings.Split(r.RemoteAddr, ":")[0]

		err := s.AddMetric(metricName, metricValue, metricType, agentIP)
		if err != nil {
			http.Error(w, "invalid metric values\n", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var reply = struct {
			Status string `json:"status"`
		}{Status: "ok"}
		body, err := json.Marshal(reply)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
		if _, err := w.Write(body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	})
}

func JSONSetterHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		var metric model.Metrics
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(&metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		err := s.AddObjMetric(metric, agentIP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return 
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var reply = struct {
			Status string `json:"status"`
		}{Status: "ok"}
		body, err := json.Marshal(reply)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
		if _, err := w.Write(body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	}
}

func MetricGetterHandler(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		reqPath := regexp.MustCompile(`^/value/(\w+)/([\w\-.]+)$`)
		matches := reqPath.FindStringSubmatch(r.URL.Path)
		if len(matches) != 3 {
			http.Error(w, "invalid path format. Use: /value/metrictype/metricname/values\n", http.StatusNotFound)
			return
		}
		metricType := matches[1]
		metricName := matches[2]
		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		
		body, err := s.GetMetric(metricName, metricType, agentIP)
		if err != nil {
			errStr := fmt.Sprintf("There was a error while getting the metric:\n%v", err)
			http.Error(w, errStr, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	})
}

func JSONGetterHandler(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusMethodNotAllowed)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		agentIP := strings.Split(r.RemoteAddr, ":")[0]
		var metricReq model.Metrics
		jsonDecoder := json.NewDecoder(r.Body)
		if err := jsonDecoder.Decode(&metricReq); err != nil {
			log.Printf("Decoding json request failed with %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricResp, err := s.GetObjMetric(metricReq, agentIP)
		if err != nil {
			log.Printf("Getting object metrics failed with %v", err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonEncoder := json.NewEncoder(w)
		if err := jsonEncoder.Encode(&metricResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func MetricsListHandler(s Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request on %v, from %v\n", r.URL.Path, r.Host)
		if r.Method != http.MethodGet {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}
		
		memStorage, metricIdx, err := s.ListMetric()
		if err != nil {
			errStr := fmt.Sprintf("There was a error while listing metrics:\n%v", err)
			http.Error(w, errStr, http.StatusInternalServerError)
			return 
		}

		body := `<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="utf-8">
		<title>Metrics</title>
		<style>
		body { font-family: sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
		th { background: #f0f0f0; }
		</style>
		</head>
		<body>
		<h1>All Metrics</h1>
		<table>
		<tr>
			<th>Agent ID</th>
			<th>Name</th>
			<th>Type</th>
			<th>Value</th>
		</tr>
		`
		for _, v := range metricIdx {

			body += "<tr>"
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(v))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(memStorage[v].ID))
			body += fmt.Sprintf("<td>%v</td>", html.EscapeString(memStorage[v].MType))
			switch memStorage[v].MType {
			case model.Counter:
				if memStorage[v].Delta != nil {
					body += fmt.Sprintf("<td>%d</td>", *memStorage[v].Delta)
				} else {
					body += "<td>-</td>"
				}
			case model.Gauge:
				if memStorage[v].Value != nil {
					body += fmt.Sprintf("<td>%f</td>", *memStorage[v].Value)
				} else {
					body += "<td>-</td>"
				}
			default:
				body += "<td>?</td>"
			}
			body += "</tr>\n"
		}
		body += `</table>
		</body>
		</html>`

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, body); err != nil {
			log.Printf("failed to write response body: %v", err)
		}
	})
}
