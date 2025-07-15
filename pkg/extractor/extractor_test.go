package extractor

import (
	"strings"
	"testing"
)

func TestExtractor_Extract(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		opts     Options
		wantLen  int
		wantErr  bool
	}{
		{
			name: "deployment with env vars",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: FOO
          value: bar
        - name: BAZ
          value: qux`,
			opts:    Options{},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "deployment with secret refs",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: SECRET_VALUE
          valueFrom:
            secretKeyRef:
              name: my-secret
              key: password`,
			opts:    Options{},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "pod with multiple containers",
			manifest: `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: app1
    env:
    - name: APP1_VAR
      value: value1
  - name: app2
    env:
    - name: APP2_VAR
      value: value2`,
			opts:    Options{Container: "app1"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "cronjob with env vars",
			manifest: `apiVersion: batch/v1
kind: CronJob
metadata:
  name: test-cronjob
spec:
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: job
            env:
            - name: JOB_TYPE
              value: batch`,
			opts:    Options{},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:     "unsupported resource",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: test-service`,
			opts:    Options{},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			reader := strings.NewReader(tt.manifest)
			
			envVars, err := e.Extract(reader, tt.opts)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && len(envVars) != tt.wantLen {
				t.Errorf("Extract() got %d env vars, want %d", len(envVars), tt.wantLen)
			}
		})
	}
}