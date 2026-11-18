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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmitDatanodeDrain(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != datanodeDrainEndpoint {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			var req drainDatanodeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}
			if req.NodeID != 7 {
				t.Fatalf("node_id = %d, want 7", req.NodeID)
			}
			_ = json.NewEncoder(w).Encode(drainDatanodeResponse{ProcedureID: "proc-1"})
		}))
		defer srv.Close()

		id, err := SubmitDatanodeDrain(srv.URL, 7)
		if err != nil {
			t.Fatalf("SubmitDatanodeDrain() error: %v", err)
		}
		if id != "proc-1" {
			t.Fatalf("SubmitDatanodeDrain() = %q, want %q", id, "proc-1")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		defer srv.Close()

		_, err := SubmitDatanodeDrain(srv.URL, 7)
		if !errors.Is(err, ErrDatanodeDrainUnsupported) {
			t.Fatalf("SubmitDatanodeDrain() error = %v, want ErrDatanodeDrainUnsupported", err)
		}
	})

	t.Run("already running", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(greptimeDBErrCodeHeader, greptimeDBErrCodeIllegalState)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(metasrvErrorBody{Code: 1006, Error: "drain is running"})
		}))
		defer srv.Close()

		_, err := SubmitDatanodeDrain(srv.URL, 7)
		if !errors.Is(err, ErrDatanodeDrainRunning) {
			t.Fatalf("SubmitDatanodeDrain() error = %v, want ErrDatanodeDrainRunning", err)
		}
	})
}

func TestProcedureStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("procedure_id") != "proc-1" {
			t.Fatalf("unexpected procedure_id: %s", r.URL.Query().Get("procedure_id"))
		}
		_ = json.NewEncoder(w).Encode(procedureStatusResponse{ProcedureID: "proc-1", Status: ProcedureStateDone})
	}))
	defer srv.Close()

	status, err := ProcedureStatus(srv.URL, "proc-1")
	if err != nil {
		t.Fatalf("ProcedureStatus() error: %v", err)
	}
	if status != ProcedureStateDone {
		t.Fatalf("ProcedureStatus() = %q, want %q", status, ProcedureStateDone)
	}
}

func TestMarkDatanodeSchedulable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datanodeMarkSchedulable {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(datanodeSchedulingResponse{NodeID: 7, Schedulable: true})
	}))
	defer srv.Close()

	if err := MarkDatanodeSchedulable(srv.URL, 7); err != nil {
		t.Fatalf("MarkDatanodeSchedulable() error: %v", err)
	}
}
