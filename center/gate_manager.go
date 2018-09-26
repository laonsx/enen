package center

import (
	"log"
	"sync"
	"time"

	"enen/common/pb"
	"github.com/laonsx/gamelib/gofunc"
	"github.com/laonsx/gamelib/timer"
	"github.com/sirupsen/logrus"
)

const (
	RefreshTime = 30
)

var GateManager *gateManager

func initGateManager() {

	GateManager = &gateManager{
		gates: make(map[string]*GateInfo),
	}

	timer.AfterFunc(time.Duration(300)*time.Second, 0, func(n int) {

		logrus.Infof("gateInfo => %v", GateManager.getGateWeights())
	})
}

type (
	GateState   int
	gateManager struct {
		sync.RWMutex
		gates map[string]*GateInfo
	}
	GateInfo struct {
		addr    string
		weight  int
		refresh int64
		state   pb.GateState
		name    string
	}
)

func (gm *gateManager) getGateWeights() map[string]int {

	currTime := time.Now().Unix()
	gates := make(map[string]int)

	gm.RLock()
	defer gm.RUnlock()

	for _, gateInfo := range gm.gates {

		if gateInfo.state != pb.GateState_Online {

			continue
		}

		if currTime > gateInfo.refresh+RefreshTime {

			gateInfo.state = pb.GateState_Offline

			log.Println("gate(%s) offline %v", gateInfo.name, gateInfo)

			continue
		}

		gates[gateInfo.name] = gateInfo.weight
	}

	return gates
}

func (gm *gateManager) getRandGateInfo() *GateInfo {

	name := gofunc.RandStrKey(gm.getGateWeights())

	gm.RLock()
	defer gm.RUnlock()

	return gm.gates[name]
}

func (gm *gateManager) getGateByName(name string) *GateInfo {

	gm.RLock()
	defer gm.RUnlock()

	return gm.gates[name]
}

func (gm *gateManager) refreshGateInfo(gateInfoPb *pb.GateInfo) {

	if gateInfoPb == nil {

		return
	}

	currTime := time.Now().Unix()

	gm.Lock()
	defer gm.Unlock()

	if _, ok := gm.gates[gateInfoPb.Name]; ok {

		if gm.gates[gateInfoPb.Name].state == pb.GateState_Close && gateInfoPb.State == pb.GateState_Close {

			return
		}

		if gm.gates[gateInfoPb.Name].state != pb.GateState_Online && gateInfoPb.State == pb.GateState_Online {

			log.Printf("gate(%s) online %v", gateInfoPb.Name, gm.gates[gateInfoPb.Name])
		}

		if gateInfoPb.State == pb.GateState_Close {

			log.Printf("gate(%s) close %v", gateInfoPb.Name, gm.gates[gateInfoPb.Name])
		}

		gm.gates[gateInfoPb.Name].refresh = currTime
		gm.gates[gateInfoPb.Name].weight = int(gateInfoPb.Weight)
		gm.gates[gateInfoPb.Name].state = gateInfoPb.State

		return
	}

	newGate := &GateInfo{
		addr:    gateInfoPb.Addr,
		weight:  int(gateInfoPb.Weight),
		refresh: currTime,
		state:   gateInfoPb.State,
		name:    gateInfoPb.Name,
	}

	gm.gates[newGate.name] = newGate

	log.Printf("gate(%s) online %v", newGate.name, newGate)
}
