/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package cdrc

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
)

// NewCDRCService instantiates the CDRCService
func NewCDRcService(cfg *config.CGRConfig, cfgRld chan string) (cdrcS *CDRCService, err error) {
	cdrcS = &CDRcService{
		rdrs:   make(map[string]CDRCReader),
		cfgRld: cfgRld,
	}
	for _, cfg := range cfg.CDRcProfiles() {
		if !cfg.Enabled {
			continue
		}

	}
	if len(cdrS.rdrs) == 0 {
		return nil, nil // no CDRC profiles enabled
	}
}

// CDRCService is managing the CDRCReaders
type CDRcService struct {
	sync.RWMutex
	cfg    *config.CGRConfig
	rdrs   map[string][]CDRcReader // list of readers on specific paths map[path]reader
	cfgRld chan struct{}           // signal the need of config reloading - chan path / *any+

}

// ListenAndServe loops keeps the service alive
func (cS *CDRcService) ListenAndServe(exitChan chan bool) error {
	go cS.handleReloads() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// cdrcCfgRef will be used to reference a specific reader
type cdrcCfgRef struct {
	path string
	idx  int
}

func (cS *CDRcService) handleReloads() {
	for {
		<-cS.cfgRld
		cfgIDs := make(map[string]*cdrcCfgRef)   // IDs which are configured in CDRcProfiles
		inUseIDs := make(map[string]*cdrcCfgRef) // IDs which are running in CDRcService indexed on path
		addIDs := make(map[string]struct{})      // IDs which need to be added to CDRcService
		remIDs := make(map[string]struct{})      // IDs which need to be removed from CDRcService
		// index config IDs
		for i, cgrCfg := range config.CDRcProfiles() {
			cfgIDs[cgrCfg.ID] = &cdrcCfgRef{path: cgrCfg.Path, idx: i}
		}
		ccS.Lock()
		// index in use IDs
		for path, rdrs := range cS.rdrs {
			for i, rdr := range rdrs {
				inUseIDs[rdr.ID()] = &cdrcCfgRef{path: path, idx: i}
			}
		}
		// find out removed ids
		for id, ref := range inUseIDs {
			if _, has := cfgIDs[id]; !has {
				remIDs[id] = struct{}{}
			}
		}
		// find out added ids
		for id, path := range cfgIDs {
			if _, has := inUseIDs[id]; !has {
				addIDs[id] = path
			}
		}
		for id := range remIDs {
			ref := inUseIDs[id]
			rdrSlc := cS.rdrs[ref.path]
			// remove the ids
			copy(rdrSlc[ref.idx:], rdrSlc[ref.idx+1:])
			rdrSlc[len(rdrSlc)-1] = nil // so it can be garbage collected
			rdrSlc = rdrSlc[:len(rdrSlc)-1]
		}
		// add new ids:
		for id := range addIDs {
			cfgRef := cfgIDs[id]
			if newRdr, err := NewCDRcReader(cS.cfg); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf(
						"<%s> error reloading config with ID: <%s>, err: <%s>",
						utils.CDRc, err.Error()))
			} else {
				cS.rdrs[path] = append(cS.rdrs[path], newRdr)
			}

		}
	}
}
