// Copyright 2026 Greptime Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/avast/retry-go"
)

const (
	datanodeDrainEndpoint    = "/admin/datanode/drain"
	datanodeMarkSchedulable  = "/admin/datanode/mark-schedulable"
	datanodeSchedulingStatus = "/admin/datanode/scheduling/status"
	procedureStatusEndpoint  = "/admin/procedure/status"

	// greptimeDBErrCodeHeader is set by metasrv on every admin API error response.
	greptimeDBErrCodeHeader = "x-greptime-err-code"

	// greptimeDBErrCodeIllegalState matches common_error::StatusCode::IllegalState (1006),
	// which metasrv returns when a datanode drain is already running.
	greptimeDBErrCodeIllegalState = "1006"
)

// Procedure states returned by the metasrv procedure status API.
const (
	ProcedureStateRunning = "Running"
	ProcedureStateDone    = "Done"
	ProcedureStateFailed  = "Failed"
	// ProcedureStatePoisoned is returned when a procedure cannot proceed anymore.
	ProcedureStatePoisoned = "Poisoned"
)

var (
	// ErrDatanodeDrainRunning indicates metasrv rejected the submission because a
	// drain for the same datanode is already running.
	ErrDatanodeDrainRunning = errors.New("datanode drain is already running")

	// ErrDatanodeDrainUnsupported indicates the metasrv does not expose the drain API,
	// e.g. on GreptimeDB versions older than v1.4.0.
	ErrDatanodeDrainUnsupported = errors.New("metasrv does not support datanode drain")
)

// DatanodeNodeID computes the GreptimeDB datanode node ID assigned by the
// initializer for a StatefulSet pod ordinal. It mirrors
// cmd/initializer/internal.ConfigGenerator.allocateNodeID:
//
//   - single datanode (groupID < 0): startNodeID + ordinal
//   - datanode group (groupID >= 0): startNodeID + (groupID<<32 | ordinal)
func DatanodeNodeID(ordinal, groupID, startNodeID int32) uint64 {
	if groupID >= 0 {
		return uint64(startNodeID) + (uint64(groupID)<<32 | uint64(ordinal))
	}
	return uint64(startNodeID) + uint64(ordinal)
}

type drainDatanodeRequest struct {
	NodeID       uint64  `json:"node_id"`
	TargetNodeID *uint64 `json:"target_node_id,omitempty"`
	Parallelism  *int    `json:"parallelism,omitempty"`
	Strategy     string  `json:"strategy,omitempty"`
}

type drainDatanodeResponse struct {
	ProcedureID string `json:"procedure_id"`
}

type datanodeRequest struct {
	NodeID uint64 `json:"node_id"`
}

type datanodeSchedulingResponse struct {
	NodeID      uint64 `json:"node_id"`
	Schedulable bool   `json:"schedulable"`
}

type procedureStatusResponse struct {
	ProcedureID string `json:"procedure_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

type metasrvErrorBody struct {
	Code  uint32 `json:"code"`
	Error string `json:"error"`
}

// SubmitDatanodeDrain submits an asynchronous drain for the datanode identified by
// nodeID and returns the procedure ID to poll. The drain fences the datanode and
// migrates every leader region off it.
func SubmitDatanodeDrain(metaHTTPServiceURL string, nodeID uint64) (string, error) {
	rsp, err := postJSON(metaHTTPServiceURL, datanodeDrainEndpoint, drainDatanodeRequest{NodeID: nodeID})
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", metasrvErrorFromResponse(rsp, "drain datanode")
	}

	var resp drainDatanodeResponse
	if err := json.NewDecoder(rsp.Body).Decode(&resp); err != nil {
		return "", err
	}
	return resp.ProcedureID, nil
}

// ProcedureStatus returns the state of a submitted procedure.
func ProcedureStatus(metaHTTPServiceURL, procedureID string) (string, error) {
	query := url.Values{"procedure_id": []string{procedureID}}.Encode()
	rsp, err := get(metaHTTPServiceURL + procedureStatusEndpoint + "?" + query)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return "", metasrvErrorFromResponse(rsp, "query procedure status")
	}

	var resp procedureStatusResponse
	if err := json.NewDecoder(rsp.Body).Decode(&resp); err != nil {
		return "", err
	}
	return resp.Status, nil
}

// MarkDatanodeSchedulable restores scheduling for the datanode identified by nodeID,
// removing the fence written by a previous drain.
func MarkDatanodeSchedulable(metaHTTPServiceURL string, nodeID uint64) error {
	rsp, err := postJSON(metaHTTPServiceURL, datanodeMarkSchedulable, datanodeRequest{NodeID: nodeID})
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return metasrvErrorFromResponse(rsp, "mark datanode schedulable")
	}
	return nil
}

// DatanodeSchedulingStatus reports whether the datanode may receive new regions.
func DatanodeSchedulingStatus(metaHTTPServiceURL string, nodeID uint64) (bool, error) {
	query := url.Values{"node_id": []string{fmt.Sprintf("%d", nodeID)}}.Encode()
	rsp, err := get(metaHTTPServiceURL + datanodeSchedulingStatus + "?" + query)
	if err != nil {
		return false, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return false, metasrvErrorFromResponse(rsp, "query datanode scheduling status")
	}

	var resp datanodeSchedulingResponse
	if err := json.NewDecoder(rsp.Body).Decode(&resp); err != nil {
		return false, err
	}
	return resp.Schedulable, nil
}

// postJSON posts the JSON-encoded payload and retries only transient network errors,
// so a business rejection from metasrv is returned as-is instead of being retried.
func postJSON(metaHTTPServiceURL, endpoint string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var rsp *http.Response
	err = retry.Do(
		func() error {
			var err error
			rsp, err = http.Post(metaHTTPServiceURL+endpoint, "application/json", bytes.NewReader(body))
			return err
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
	)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func get(requestURL string) (*http.Response, error) {
	var rsp *http.Response
	err := retry.Do(
		func() error {
			var err error
			rsp, err = http.Get(requestURL)
			return err
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
	)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func metasrvErrorFromResponse(rsp *http.Response, action string) error {
	// The drain endpoint does not exist before GreptimeDB v1.4.0.
	if rsp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s (status %d)", ErrDatanodeDrainUnsupported, action, rsp.StatusCode)
	}

	if rsp.Header.Get(greptimeDBErrCodeHeader) == greptimeDBErrCodeIllegalState {
		return fmt.Errorf("%w: %s", ErrDatanodeDrainRunning, action)
	}

	msg := ""
	data, _ := io.ReadAll(io.LimitReader(rsp.Body, 4096))
	var body metasrvErrorBody
	if json.Unmarshal(data, &body) == nil && body.Error != "" {
		msg = body.Error
	} else {
		msg = strings.TrimSpace(string(data))
	}
	if msg == "" {
		msg = rsp.Status
	}
	return fmt.Errorf("%s failed with status %d: %s", action, rsp.StatusCode, msg)
}
