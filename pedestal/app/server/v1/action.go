package v1

import (
	"strconv"
	"strings"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/log"

	"github.com/gogf/gf/v2/os/glog"
)

type ListActionResV1 struct {
	Actions []*ActionResV1 `json:"actions"`
	Total   int            `json:"total"`
}

type ActionResV1 struct {
	Name        string  `json:"name"`
	Version     float32 `json:"version"`
	Description string  `json:"description"`
}

func ListAction() (*ListActionResV1, error) {
	stage := log.NewStage().Enter("ListAction")

	actions := action.ListActionKey(stage)
	res := &ListActionResV1{
		Actions: []*ActionResV1{},
		Total:   len(actions),
	}

	for _, key := range actions {
		ns := strings.Split(key, "@")
		if len(ns) != 2 {
			glog.Warningf(stage.Context(), "format action[%v] key failed: len(%v)!=2", key, len(ns))
			continue
		}
		ff, err := strconv.ParseFloat(ns[1], 10)
		if err != nil {
			glog.Warningf(stage.Context(), "format action[%v] key failed: %v", key, err)
			continue
		}

		a, _ := action.GetAction(stage, ns[0], float32(ff))
		res.Actions = append(res.Actions, &ActionResV1{
			Name:        a.Name(),
			Version:     a.Version(),
			Description: a.Description(),
		})
	}
	return res, nil
}
