package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	pluginNameTablePrefix = "{{cookiecutter.plugin_id}}-table-"
)

func setupSocket(socketPath string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socketPath))
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socketPath), err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socketPath, err)
	}

	log.Printf("Listening on: unix://%s", socketPath)
	return listener, nil
}

func setupSignals(socketPath string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.RemoveAll(filepath.Dir(socketPath))
		os.Exit(0)
	}()
}

func main() {
	// We put the socket in a sub-directory to have more control on the permissions
	const socketPath = "/var/run/scope/plugins/{{cookiecutter.plugin_id}}/{{cookiecutter.plugin_id}}.sock"
	hostID, _ := os.Hostname()

	// Handle the exit signal
	setupSignals(socketPath)

	log.Printf("Starting on %s...\n", hostID)

	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	plugin := &Plugin{HostID: hostID}
	plugin.stateControlA = true;
	plugin.stateControlB = false;

	// The /report endpoint is a mandatory interface, every plugin MUST implement it
	http.HandleFunc("/report", plugin.Report)

	// The /control endpoint is optionals and allow plugin to provided to the user actions
	http.HandleFunc("/control", plugin.Control)

	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Plugin groups the methods a plugin needs
type Plugin struct {
	lock sync.Mutex

	HostID string

	stateControlA bool
	stateControlB bool
}

type request struct {
	NodeID  string
	Control string
}

type response struct {
	ShortcutReport *report `json:"shortcutReport,omitempty"`
}

type report struct {
	Host    topology
	Plugins []pluginSpec
}

// This is a generic representation of a topology
// Some examples of topology are : Host, Container, Pod , Service, Deployment , ReplicaSet
type topology struct {
	Nodes    map[string]node    `json:"nodes"`
	Controls map[string]control `json:"controls"`

	MetadataTemplates map[string]metadataTemplate `json:"metadata_templates,omitempty"`
	TableTemplates    map[string]tableTemplate    `json:"table_templates,omitempty"`
	MetricTemplates   map[string]metricTemplate   `json:"metric_templates"`
}

// This structure describe a table
// the prefix is used to match the row to the right table
type tableTemplate struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
}

type metadataTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`    // Human-readable descriptor for this row
	Truncate int     `json:"truncate,omitempty"` // If > 0, truncate the value to this length.
	Datatype string  `json:"dataType,omitempty"`
	Priority float64 `json:"priority,omitempty"`
	From     string  `json:"from,omitempty"` // Defines how to get the value from a report node
}

// This describe a data of metric type
// Metrics are visualized as a graph on the UI (see https://github.com/weaveworks-plugins/scope-http-statistics)
type metricTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`
	Format   string  `json:"format,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

// The actual data for the metric visualization
type metric struct {
	Samples []sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

// A metric sample
type sample struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type node struct {
	Metrics        map[string]metric       `json:"metrics"`
	LatestControls map[string]controlEntry `json:"latestControls,omitempty"`
	Latest         map[string]stringEntry  `json:"latest,omitempty"`
}

type controlEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     controlData `json:"value"`
}

type controlData struct {
	Dead bool `json:"dead"`
}

type stringEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"`
	Rank  int    `json:"rank"`
}

// Plugin specification
type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

// This function makes the report
func (p *Plugin) makeReport() (*report, error) {
	timestamp := time.Now()
	metrics, err := p.metrics()
	if err != nil {
		return nil, err
	}
	rpt := &report{
		Host: topology{
			Nodes: map[string]node{
				p.getTopologyHost(): {
					Metrics:        metrics,
					LatestControls: p.latestControls(),
					Latest:        	map[string]stringEntry{
						fmt.Sprintf("%s%s", "{{cookiecutter.plugin_id}}-", "label-1"): {
							Timestamp: timestamp,
							Value:     "Value 1",
						},
						fmt.Sprintf("%s%s", pluginNameTablePrefix, "label-2"): {
							Timestamp: timestamp,
							Value:     "Value 2",
						},
					},
				},
			},
			MetricTemplates: p.metricTemplates(),
			MetadataTemplates: getMetadataTemplate(),
			TableTemplates:    getTableTemplate(),
			Controls:        p.controls(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "{{cookiecutter.plugin_id}}",
				Label:       "{{cookiecutter.plugin_name}}",
				Description: "{{cookiecutter.plugin_description}}",
				Interfaces:  []string{"reporter", "controller"},
				APIVersion:  "1",
			},
		},
	}
	return rpt, nil
}

func (p *Plugin) metrics() (map[string]metric, error) {
	value, err := p.metricValue()
	if err != nil {
		return nil, err
	}
	id, _ := p.metricIDAndName()
	metrics := map[string]metric{
		id: {
			Samples: []sample{
				{
					Date:  time.Now(),
					Value: value,
				},
			},
			Min: 0,
			Max: 100,
		},
	}
	return metrics, nil
}

func (p *Plugin) latestControls() map[string]controlEntry {
	ts := time.Now()
	ctrls := map[string]controlEntry{}
	for _, details := range p.allControlDetails() {
		ctrls[details.id] = controlEntry{
			Timestamp: ts,
			Value: controlData{
				Dead: details.dead,
			},
		}
	}
	return ctrls
}

func (p *Plugin) metricTemplates() map[string]metricTemplate {
	id, name := p.metricIDAndName()
	return map[string]metricTemplate{
		id: {
			ID:       id,
			Label:    name,
			Format:   "percent",
			Priority: 0.1,
		},
	}
}

// This returns all the available controls
func (p *Plugin) controls() map[string]control {
	ctrls := map[string]control{}
	for _, details := range p.allControlDetails() {
		ctrls[details.id] = control{
			ID:    details.id,
			Human: details.human,
			Icon:  details.icon,
			Rank:  1,
		}
	}
	return ctrls
}

// Report is the handler of the "/report" endpoint.
// It implements the "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Println(r.URL.String())
	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raw, err := json.Marshal(*rpt)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// Control is the handler of the "/control" endpoint.
// It implements the "controller" interface.
func (p *Plugin) Control(w http.ResponseWriter, r *http.Request) {
	p.lock.Lock()
	defer p.lock.Unlock()
	log.Println(r.URL.String())
	xreq := request{}
	err := json.NewDecoder(r.Body).Decode(&xreq)
	if err != nil {
		log.Printf("Bad request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	thisNodeID := p.getTopologyHost()
	if xreq.NodeID != thisNodeID {
		log.Printf("Bad nodeID, expected %q, got %q", thisNodeID, xreq.NodeID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedControlID, _, _, _ := p.controlDetails()
	if expectedControlID != xreq.Control {
		log.Printf("Bad control, expected %q, got %q", expectedControlID, xreq.Control)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.stateControlA = !p.stateControlA
	p.stateControlB = !p.stateControlB

	rpt, err := p.makeReport()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := response{ShortcutReport: rpt}
	raw, err := json.Marshal(res)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// getTopologyHost prepares node value for the host topology
func (p *Plugin) getTopologyHost() string {
	return fmt.Sprintf("%s;<host>", p.HostID)
}

func getMetadataTemplate() map[string]metadataTemplate {
	return map[string]metadataTemplate{
		"{{cookiecutter.plugin_id}}-label-1": {
			ID:       "{{cookiecutter.plugin_id}}-label-1",
			Label:    "Label 1",
			Truncate: 0,
			Datatype: "",
			Priority: 13.5,
			From:     "latest",
		},
		"{{cookiecutter.plugin_id}}-label-2": {
			ID:       "{{cookiecutter.plugin_id}}-label-2",
			Label:    "Label 2",
			Truncate: 0,
			Datatype: "",
			Priority: 13.6,
			From:     "latest",
		},
	}
}

func getTableTemplate() map[string]tableTemplate {
	return map[string]tableTemplate{
		"{{cookiecutter.plugin_id}}-table": {
			ID:     "{{cookiecutter.plugin_id}}-table",
			Label:  "Plugin Table",
			Prefix: pluginNameTablePrefix,
		},
	}
}

func (p *Plugin) metricIDAndName() (string, string) {
	return "metric-id", "matric-name"
}

// metricValue returns the value of the metric
func (p *Plugin) metricValue() (float64, error) {
	return 0.42, nil
}

type controlDetails struct {
	id    string
	human string
	icon  string
	rank  float32
	dead  bool
}

func (p *Plugin) allControlDetails() []controlDetails {
	return []controlDetails{
		{
			id:    "control-a",
			human: "Contorl A",
			icon:  "fa-bomb",
			rank:  1,
			dead:  p.stateControlA,
		},
		{
			id:    "control-b",
			human: "Contorl B",
			icon:  "fa-adjust",
			rank:  1.5,
			dead:  p.stateControlB,
		},
	}
}

func (p *Plugin) controlDetails() (string, string, string, float32) {
	for _, details := range p.allControlDetails() {
		if !details.dead {
			return details.id, details.human, details.icon, details.rank
		}
	}
	return "", "", "", 0.0
}
