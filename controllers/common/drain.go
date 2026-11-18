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

// httpResponse is the parts of an HTTP response the drain client needs after the
// body has been fully read and closed.
type httpResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

// SubmitDatanodeDrain submits an asynchronous drain for the datanode identified by
// nodeID and returns the procedure ID to poll. The drain fences the datanode and
// migrates every leader region off it.
func SubmitDatanodeDrain(metaHTTPServiceURL string, nodeID uint64) (string, error) {
	resp, err := postJSON(metaHTTPServiceURL, datanodeDrainEndpoint, drainDatanodeRequest{NodeID: nodeID})
	if err != nil {
		return "", err
	}

	if resp.statusCode != http.StatusOK {
		return "", metasrvErrorFromResponse(resp, "drain datanode")
	}

	var result drainDatanodeResponse
	if err := json.Unmarshal(resp.body, &result); err != nil {
		return "", err
	}
	return result.ProcedureID, nil
}

// ProcedureStatus returns the state of a submitted procedure.
func ProcedureStatus(metaHTTPServiceURL, procedureID string) (string, error) {
	query := url.Values{"procedure_id": []string{procedureID}}.Encode()
	resp, err := httpGet(metaHTTPServiceURL + procedureStatusEndpoint + "?" + query)
	if err != nil {
		return "", err
	}

	if resp.statusCode != http.StatusOK {
		return "", metasrvErrorFromResponse(resp, "query procedure status")
	}

	var result procedureStatusResponse
	if err := json.Unmarshal(resp.body, &result); err != nil {
		return "", err
	}
	return result.Status, nil
}

// MarkDatanodeSchedulable restores scheduling for the datanode identified by nodeID,
// removing the fence written by a previous drain.
func MarkDatanodeSchedulable(metaHTTPServiceURL string, nodeID uint64) error {
	resp, err := postJSON(metaHTTPServiceURL, datanodeMarkSchedulable, datanodeRequest{NodeID: nodeID})
	if err != nil {
		return err
	}

	if resp.statusCode != http.StatusOK {
		return metasrvErrorFromResponse(resp, "mark datanode schedulable")
	}
	return nil
}

// DatanodeSchedulingStatus reports whether the datanode may receive new regions.
func DatanodeSchedulingStatus(metaHTTPServiceURL string, nodeID uint64) (bool, error) {
	query := url.Values{"node_id": []string{fmt.Sprintf("%d", nodeID)}}.Encode()
	resp, err := httpGet(metaHTTPServiceURL + datanodeSchedulingStatus + "?" + query)
	if err != nil {
		return false, err
	}

	if resp.statusCode != http.StatusOK {
		return false, metasrvErrorFromResponse(resp, "query datanode scheduling status")
	}

	var result datanodeSchedulingResponse
	if err := json.Unmarshal(resp.body, &result); err != nil {
		return false, err
	}
	return result.Schedulable, nil
}

// postJSON posts the JSON-encoded payload and retries only transient network errors,
// so a business rejection from metasrv is returned as-is instead of being retried.
// The response body is fully read and closed before returning.
func postJSON(metaHTTPServiceURL, endpoint string, payload any) (*httpResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var resp httpResponse
	err = retry.Do(
		func() error {
			rsp, err := http.Post(metaHTTPServiceURL+endpoint, "application/json", bytes.NewReader(body))
			if err != nil {
				return err
			}
			defer rsp.Body.Close()

			resp.statusCode = rsp.StatusCode
			resp.header = rsp.Header
			resp.body, err = io.ReadAll(rsp.Body)
			return err
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
	)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// httpGet fetches the given URL and retries only transient network errors. The
// response body is fully read and closed before returning.
func httpGet(requestURL string) (*httpResponse, error) {
	var resp httpResponse
	err := retry.Do(
		func() error {
			rsp, err := http.Get(requestURL)
			if err != nil {
				return err
			}
			defer rsp.Body.Close()

			resp.statusCode = rsp.StatusCode
			resp.header = rsp.Header
			resp.body, err = io.ReadAll(rsp.Body)
			return err
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
	)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func metasrvErrorFromResponse(resp *httpResponse, action string) error {
	// The drain endpoints do not exist before GreptimeDB v1.4.0.
	if resp.statusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s (status %d)", ErrDatanodeDrainUnsupported, action, resp.statusCode)
	}

	if resp.header.Get(greptimeDBErrCodeHeader) == greptimeDBErrCodeIllegalState {
		return fmt.Errorf("%w: %s", ErrDatanodeDrainRunning, action)
	}

	msg := ""
	var body metasrvErrorBody
	if json.Unmarshal(resp.body, &body) == nil && body.Error != "" {
		msg = body.Error
	} else {
		msg = strings.TrimSpace(string(resp.body))
	}
	if msg == "" {
		msg = http.StatusText(resp.statusCode)
	}
	return fmt.Errorf("%s failed with status %d: %s", action, resp.statusCode, msg)
}
