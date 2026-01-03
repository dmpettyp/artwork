package imagegen

import "time"

func (ig *ImageGen) observeTotal(nodeType string, start time.Time, err error) {
	if ig.metrics == nil {
		return
	}
	status := "success"
	if err != nil {
		status = "error"
	}
	ig.metrics.ObserveTotal(nodeType, status, time.Since(start))
}

type imageGenMetricsRecorder struct {
	ig       *ImageGen
	nodeType string
	start    time.Time
}

func (ig *ImageGen) newRecorder(nodeType string) *imageGenMetricsRecorder {
	return &imageGenMetricsRecorder{
		ig:       ig,
		nodeType: nodeType,
		start:    time.Now(),
	}
}

func (r *imageGenMetricsRecorder) preview(err error) {
	if r.ig.metrics == nil {
		return
	}
	status := "success"
	if err != nil {
		status = "error"
	}
	r.ig.metrics.ObservePreview(r.nodeType, status)
}

func (r *imageGenMetricsRecorder) output(err error) {
	if r.ig.metrics == nil {
		return
	}
	status := "success"
	if err != nil {
		status = "error"
	}
	r.ig.metrics.ObserveOutput(r.nodeType, status)
}

func (r *imageGenMetricsRecorder) total(err error) {
	r.ig.observeTotal(r.nodeType, r.start, err)
}
